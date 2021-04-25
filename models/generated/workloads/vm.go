package workloads

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net"

	"github.com/rs/zerolog/log"
	schema "github.com/threefoldtech/tfexplorer/schema"
)

var _ Workloader = (*VirtualMachine)(nil)
var _ Capaciter = (*VirtualMachine)(nil)

type VirtualMachine struct {
	ReservationInfo `bson:",inline"`

	Flist     string   `bson:"flist" json:"flist"`
	NetworkId string   `bson:"network_id" json:"network_id"`
	Ipaddress net.IP   `bson:"ipaddress" json:"ipaddress"`
	SshKeys   []string `bson:"ssh_keys" json:"ssh_keys"`
	// why isn't this a net.IP? because it's a wid
	PublicIP schema.ID  `bson:"public_ip" json:"public_ip"`
	Capacity VMCapacity `bson:"capcity" json:"capacity"`
}

func (vm *VirtualMachine) GetRSU() (RSU, error) {
	return vm.Capacity.GetRSU()
}

func (vm *VirtualMachine) SignatureChallenge() ([]byte, error) {
	log.Info().Msg("Entering signature challenge")
	ric, err := vm.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}
	log.Info().Msg(vm.Flist)

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", vm.Flist); err != nil {
		return nil, err
	}
	log.Info().Msg(vm.NetworkId)
	if _, err := fmt.Fprintf(b, "%s", vm.NetworkId); err != nil {
		return nil, err
	}
	log.Info().Msg(fmt.Sprint(vm.PublicIP))
	if _, err := fmt.Fprintf(b, "%d", vm.PublicIP); err != nil {
		return nil, err
	}
	log.Info().Msg(vm.Ipaddress.String())

	if _, err := fmt.Fprintf(b, "%s", vm.Ipaddress.String()); err != nil {
		return nil, err
	}

	for _, key := range vm.SshKeys {
		log.Info().Msg(key)
		if _, err := fmt.Fprintf(b, "%s", key); err != nil {
			return nil, err
		}
	}

	if err := vm.Capacity.SigningEncode(b); err != nil {
		return nil, err
	}
	log.Info().Msg("Entering signature 7")
	log.Info().Bytes("customer data", b.Bytes()).Msg("Hi")
	return b.Bytes(), nil
}

type VMCapacity struct {
	Cpu      int64  `bson:"cpu" json:"cpu"`
	Memory   int64  `bson:"memory" json:"memory"`
	DiskSize uint64 `bson:"disk_size" json:"disk_size"`
}

func (vm VMCapacity) SigningEncode(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%d", vm.Cpu); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%d", vm.Memory); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%d", vm.DiskSize); err != nil {
		return err
	}
	return nil
}

func (vm VMCapacity) GetRSU() (RSU, error) {
	rsu := RSU{
		CRU: vm.Cpu,
		// round mru to 4 digits precision
		MRU: math.Round(float64(vm.Memory)/1024*10000) / 10000,
	}
	storageSize := math.Round(float64(vm.DiskSize)/1024*10000) / 10000
	storageSize = math.Max(0, storageSize-50) // we offer the 50 first GB of storage for container root
	rsu.SRU = storageSize
	return rsu, nil
}
