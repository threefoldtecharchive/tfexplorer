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
}

func (z *PublicIP) GetRSU() RSU {
	return RSU{IPV4U: 1}
}

func (z *PublicIP) SignatureChallenge() ([]byte, error) {
	ric, err := z.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer(ric)
	fmt.Fprintf(b, "%v", z.IPaddress)

	return b.Bytes(), nil
}
