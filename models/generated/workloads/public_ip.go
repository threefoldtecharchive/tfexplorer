package workloads

import (
	"bytes"
	"fmt"
	"net"
)

var _ Workloader = (*PublicIP)(nil)
var _ Capaciter = (*PublicIP)(nil)

// PublicIP is a struct that defines the workload to reserve a public ip on the grid
type PublicIP struct {
	ReservationInfo `bson:",inline"`

	IP net.IP `bson:"ip" json:"ip"`
}

func (z *PublicIP) GetRSU() RSU {
	return RSU{}
}

func (z *PublicIP) SignatureChallenge() ([]byte, error) {
	ric, err := z.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer(ric)
	fmt.Fprintf(b, "%v", z.IP)

	return b.Bytes(), nil
}
