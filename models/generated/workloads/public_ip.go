package workloads

import "net"

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
