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

	net.IP        `bson:"ip" json:"ip"`
	DestinationIP net.IP `bson:"destination_ip" json:"destination_ip"`
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
	fmt.Fprintf(b, "%v", z.DestinationIP)

	return b.Bytes(), nil
}
