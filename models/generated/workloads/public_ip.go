package workloads

import (
	"bytes"
	"fmt"

	"github.com/threefoldtech/tfexplorer/schema"
)

var _ Workloader = (*PublicIP)(nil)
var _ Capaciter = (*PublicIP)(nil)

// PublicIP is a struct that defines the workload to reserve a public ip on the grid
type PublicIP struct {
	ReservationInfo `bson:",inline"`

	IPaddress schema.IPCidr `bson:"ipaddress" json:"ipaddress"`
	NrName    string        `bson:"nr_name" json:"nr_name"`
}

func (z *PublicIP) GetRSU() (RSU, error) {
	return RSU{IPV4U: 1}, nil
}

func (z *PublicIP) SignatureChallenge() ([]byte, error) {
	ric, err := z.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer(ric)
	fmt.Fprintf(b, "%v", z.IPaddress)
	fmt.Fprintf(b, "%s", z.NrName)

	return b.Bytes(), nil
}
