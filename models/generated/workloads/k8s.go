package workloads

import (
	"bytes"
	"fmt"
	"net"
)

var _ Workloader = (*K8S)(nil)
var _ Capaciter = (*K8S)(nil)

type K8S struct {
	ReservationInfo `bson:",inline"`

	Size            int64             `bson:"size" json:"size"`
	ClusterSecret   string            `bson:"cluster_secret" json:"cluster_secret"`
	NetworkId       string            `bson:"network_id" json:"network_id"`
	Ipaddress       net.IP            `bson:"ipaddress" json:"ipaddress"`
	MasterIps       []net.IP          `bson:"master_ips" json:"master_ips"`
	SshKeys         []string          `bson:"ssh_keys" json:"ssh_keys"`
	StatsAggregator []StatsAggregator `bson:"stats_aggregator" json:"stats_aggregator"`
}

func (k *K8S) GetRSU() RSU {
	switch k.Size {
	case 1:
		return RSU{
			CRU: 1,
			MRU: 2,
			SRU: 50,
		}
	case 2:
		return RSU{
			CRU: 2,
			MRU: 4,
			SRU: 100,
		}
	case 3:
		return RSU{
			CRU: 2,
			MRU: 8,
			SRU: 25,
		}
	case 4:
		return RSU{
			CRU: 2,
			MRU: 5,
			SRU: 50,
		}
	case 5:
		return RSU{
			CRU: 2,
			MRU: 8,
			SRU: 200,
		}
	case 6:
		return RSU{
			CRU: 4,
			MRU: 16,
			SRU: 50,
		}
	case 7:
		return RSU{
			CRU: 4,
			MRU: 16,
			SRU: 100,
		}
	case 8:
		return RSU{
			CRU: 4,
			MRU: 16,
			SRU: 400,
		}
	case 9:
		return RSU{
			CRU: 8,
			MRU: 32,
			SRU: 100,
		}
	case 10:
		return RSU{
			CRU: 8,
			MRU: 32,
			SRU: 200,
		}
	case 11:
		return RSU{
			CRU: 8,
			MRU: 32,
			SRU: 800,
		}
	case 12:
		return RSU{
			CRU: 1,
			MRU: 64,
			SRU: 200,
		}
	case 13:
		return RSU{
			CRU: 1,
			MRU: 64,
			SRU: 400,
		}
	case 14:
		return RSU{
			CRU: 1,
			MRU: 64,
			SRU: 800,
		}
	}
	return RSU{}
}

func (k *K8S) SignatureChallenge() ([]byte, error) {
	ric, err := k.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%d", k.Size); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", k.ClusterSecret); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", k.NetworkId); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", k.Ipaddress.String()); err != nil {
		return nil, err
	}
	for _, ip := range k.MasterIps {
		if _, err := fmt.Fprintf(b, "%s", ip.String()); err != nil {
			return nil, err
		}
	}
	for _, key := range k.SshKeys {
		if _, err := fmt.Fprintf(b, "%s", key); err != nil {
			return nil, err
		}
	}

	return b.Bytes(), nil
}
