package workloads

import (
	"encoding/json"
	"net"
	"reflect"

	"github.com/pkg/errors"
)

var _ Workloader = (*K8S)(nil)
var _ Capaciter = (*K8S)(nil)

type K8S struct {
	ReservationInfo `bson:",inline"`

	Size            int64             `bson:"size" json:"size"`
	NetworkId       string            `bson:"network_id" json:"network_id"`
	Ipaddress       net.IP            `bson:"ipaddress" json:"ipaddress"`
	ClusterSecret   string            `bson:"cluster_secret" json:"cluster_secret"`
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
	}
	return RSU{}
}

func (v *K8S) VerifyJSON() error {
	dup := K8S{}

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
