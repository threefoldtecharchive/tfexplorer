package workloads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"sort"
)

var _ Workloader = (*Container)(nil)
var _ Capaciter = (*Container)(nil)

type Container struct {
	ReservationInfo `bson:",inline"`

	Flist             string              `bson:"flist" json:"flist"`
	HubUrl            string              `bson:"hub_url" json:"hub_url"`
	Environment       map[string]string   `bson:"environment" json:"environment"`
	SecretEnvironment map[string]string   `bson:"secret_environment" json:"secret_environment"`
	Entrypoint        string              `bson:"entrypoint" json:"entrypoint"`
	Interactive       bool                `bson:"interactive" json:"interactive"`
	Volumes           []ContainerMount    `bson:"volumes" json:"volumes"`
	NetworkConnection []NetworkConnection `bson:"network_connection" json:"network_connection"`
	Stats             []Stats             `bson:"stats" json:"stats"`
	Logs              []Logs              `bson:"logs" json:"logs"`
	Capacity          ContainerCapacity   `bson:"capcity" json:"capacity"`
}

func (c *Container) GetRSU() RSU {
	return c.Capacity.GetRSU()
}

func (c *Container) SignatureChallenge() ([]byte, error) {
	ric, err := c.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	encodeEnv := func(w io.Writer, env map[string]string) error {

		keys := make([]string, 0, len(env))
		for k := range env {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			if _, err := fmt.Fprintf(w, "%s=%s", k, env[k]); err != nil {
				return err
			}
		}

		return nil
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", c.Flist); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", c.HubUrl); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", c.Entrypoint); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%t", c.Interactive); err != nil {
		return nil, err
	}
	if err := encodeEnv(b, c.Environment); err != nil {
		return nil, err
	}
	if err := encodeEnv(b, c.SecretEnvironment); err != nil {
		return nil, err
	}
	for _, v := range c.Volumes {
		if err := v.SigningEncode(b); err != nil {
			return nil, err
		}
	}
	for _, v := range c.NetworkConnection {
		if err := v.SigningEncode(b); err != nil {
			return nil, err
		}
	}

	if err := c.Capacity.SigningEncode(b); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

type ContainerCapacity struct {
	Cpu      int64        `bson:"cpu" json:"cpu"`
	Memory   int64        `bson:"memory" json:"memory"`
	DiskSize uint64       `bson:"disk_size" json:"disk_size"`
	DiskType DiskTypeEnum `bson:"disk_type" json:"disk_type"`
}

func (c ContainerCapacity) SigningEncode(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%d", c.Cpu); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%d", c.Memory); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%d", c.DiskSize); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s", c.DiskType.String()); err != nil {
		return err
	}
	return nil
}

func (c ContainerCapacity) GetRSU() RSU {
	rsu := RSU{
		CRU: c.Cpu,
		// round mru to 4 digits precision
		MRU: math.Round(float64(c.Memory)/1024*10000) / 10000,
	}
	storageSize := math.Round(float64(c.DiskSize)/1024*10000) / 10000
	storageSize = math.Max(0, storageSize-50) // we offer the 50 first GB of storage for container root
	switch c.DiskType {
	case DiskTypeHDD:
		rsu.HRU = storageSize
	case DiskTypeSSD:
		rsu.SRU = storageSize
	}

	return rsu
}

type Logs struct {
	Type string    `bson:"type" json:"type"`
	Data LogsRedis `bson:"data" json:"data"`
}

type Stats struct {
	Type string          `bson:"type" json:"type"`
	Data json.RawMessage `bson:"data" json:"data"`
}

func (c Logs) SigningEncode(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s", c.Type); err != nil {
		return err
	}
	if err := c.Data.SigningEncode(w); err != nil {
		return err
	}
	return nil
}

type LogsRedis struct {
	Stdout string `bson:"stdout" json:"stdout"`
	Stderr string `bson:"stderr" json:"stderr"`

	// Same as stdout, stderr urls but encrypted
	// with the node public key.
	SecretStdout string `bson:"secret_stdout" json:"secret_stdout"`
	SecretStderr string `bson:"secret_stderr" json:"secret_stderr"`
}

type StatsRedis struct {
	Endpoint string `bson:"endpoint" json:"endpoint"`
}

func (l LogsRedis) SigningEncode(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s", l.Stdout); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s", l.Stderr); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "%s", l.SecretStdout); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s", l.SecretStderr); err != nil {
		return err
	}
	return nil
}

type ContainerMount struct {
	VolumeId   string `bson:"volume_id" json:"volume_id"`
	Mountpoint string `bson:"mountpoint" json:"mountpoint"`
}

func (c ContainerMount) SigningEncode(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s", c.VolumeId); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s", c.Mountpoint); err != nil {
		return err
	}
	return nil
}

type NetworkConnection struct {
	NetworkId   string `bson:"network_id" json:"network_id"`
	Ipaddress   net.IP `bson:"ipaddress" json:"ipaddress"`
	PublicIp6   bool   `bson:"public_ip6" json:"public_ip6"`
	YggdrasilIP bool   `bson:"yggdrasil_ip" json:"yggdrasil_ip"`
}

func (n NetworkConnection) SigningEncode(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s", n.NetworkId); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s", n.Ipaddress.String()); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%t", n.PublicIp6); err != nil {
		return err
	}
	// TODO: re-enable when working on https://github.com/threefoldtech/zos/issues/868
	// if _, err := fmt.Fprintf(w, "%t", n.YggdrasilIP); err != nil {
	// 	return err
	// }
	return nil
}

func (s Stats) SigningEncode(w io.Writer) error {
	return nil
	// if _, err := fmt.Fprintf(w, "%s", s.Type); err != nil {
	// 	return err
	// }
	// if _, err := fmt.Fprintf(w, "%s", s.Data.Endpoint); err != nil {
	// 	return err
	// }
	// return nil
}
