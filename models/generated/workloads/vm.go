package workloads

import (
	"bytes"
	"fmt"
	"net"

	schema "github.com/threefoldtech/tfexplorer/schema"
)

var _ Workloader = (*VirtualMachine)(nil)
var _ Capaciter = (*VirtualMachine)(nil)

type VMCustomSize struct {
	CRU int64   `bson:"cru" json:"cru" `
	MRU float64 `bson:"mru" json:"mru" `
	SRU float64 `bson:"sru" json:"sru" `
}

type VirtualMachine struct {
	ReservationInfo `bson:",inline"`

	Name      string   `bson:"name" json:"name"`
	NetworkId string   `bson:"network_id" json:"network_id"`
	Ipaddress net.IP   `bson:"ipaddress" json:"ipaddress"`
	SshKeys   []string `bson:"ssh_keys" json:"ssh_keys"`
	// why isn't this a net.IP? because it's a wid
	PublicIP   schema.ID    `bson:"public_ip" json:"public_ip"`
	Size       int64        `bson:"size" json:"size"`
	CustomSize VMCustomSize `bson:"custom_size" json:"custom_size"`
}

func (k *VirtualMachine) GetRSU() (RSU, error) {
	if k.Size == -1 {
		return RSU{
			CRU: k.CustomSize.CRU,
			MRU: k.CustomSize.MRU,
			SRU: k.CustomSize.SRU,
		}, nil
	}

	rsu, ok := k8sSize[k.Size]
	if !ok {
		return RSU{}, fmt.Errorf("VM size %d is not supported", k.Size)
	}
	return rsu, nil
}

func (vm *VirtualMachine) SignatureChallenge() ([]byte, error) {
	ric, err := vm.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", vm.Name); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", vm.NetworkId); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%d", vm.PublicIP); err != nil {
		return nil, err
	}

	if _, err := fmt.Fprintf(b, "%s", vm.Ipaddress.String()); err != nil {
		return nil, err
	}

	for _, key := range vm.SshKeys {
		if _, err := fmt.Fprintf(b, "%s", key); err != nil {
			return nil, err
		}
	}

	if _, err := fmt.Fprintf(b, "%d", vm.Size); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
