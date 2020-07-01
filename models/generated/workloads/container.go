package workloads

import (
	"encoding/json"
	"math"
	"net"
	"reflect"

	"github.com/pkg/errors"
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
	StatsAggregator   []StatsAggregator   `bson:"stats_aggregator" json:"stats_aggregator"`
	Logs              []Logs              `bson:"logs" json:"logs"`
	Capacity          ContainerCapacity   `bson:"capcity" json:"capacity"`
}

func (c *Container) GetRSU() RSU {
	rsu := RSU{
		CRU: c.Capacity.Cpu,
		// round mru to 4 digits precision
		MRU: math.Round(float64(c.Capacity.Memory)/1024*10000) / 10000,
	}
	switch c.Capacity.DiskType {
	case DiskTypeHDD:
		hru := math.Round(float64(c.Capacity.DiskSize)/1024*10000) / 10000
		rsu.HRU = int64(hru)
	case DiskTypeSSD:
		sru := math.Round(float64(c.Capacity.DiskSize)/1024*10000) / 10000
		rsu.SRU = int64(sru)
	}

	return rsu
}

func (v *Container) VerifyJSON() error {
	dup := Container{}

	if err := json.Unmarshal([]byte(v.Json), &dup); err != nil {
		return errors.Wrap(err, "invalid json data")
	}

	// override the fields which are not part of the signature
	dup.ID = v.ID
	dup.Json = v.Json
	dup.CustomerTid = v.CustomerTid
	dup.NextAction = v.NextAction
	dup.SignaturesProvision = v.SignaturesProvision
	dup.SignatureFarmer = v.SignatureFarmer
	dup.SignaturesDelete = v.SignaturesDelete
	dup.Epoch = v.Epoch
	dup.Metadata = v.Metadata
	dup.Result = v.Result
	dup.WorkloadType = v.WorkloadType

	if match := reflect.DeepEqual(v, dup); !match {
		return errors.New("json data does not match actual data")
	}

	return nil
}

type ContainerCapacity struct {
	Cpu      int64        `bson:"cpu" json:"cpu"`
	Memory   int64        `bson:"memory" json:"memory"`
	DiskSize uint64       `bson:"disk_size" json:"disk_size"`
	DiskType DiskTypeEnum `bson:"disk_type" json:"disk_type"`
}

type Logs struct {
	Type string    `bson:"type" json:"type"`
	Data LogsRedis `bson:"data" json:"data"`
}

type LogsRedis struct {
	Stdout string `bson:"stdout" json:"stdout"`
	Stderr string `bson:"stderr" json:"stderr"`
}

type ContainerMount struct {
	VolumeId   string `bson:"volume_id" json:"volume_id"`
	Mountpoint string `bson:"mountpoint" json:"mountpoint"`
}

type NetworkConnection struct {
	NetworkId string `bson:"network_id" json:"network_id"`
	Ipaddress net.IP `bson:"ipaddress" json:"ipaddress"`
	PublicIp6 bool   `bson:"public_ip6" json:"public_ip6"`
}

type StatsAggregator struct {
	Type string     `bson:"type" json:"type"`
	Data StatsRedis `bson:"data" json:"data"`
}

type StatsRedis struct {
	Endpoint string `bson:"stdout" json:"endpoint"`
}
