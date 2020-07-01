package workloads

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
	schema "github.com/threefoldtech/tfexplorer/schema"
)

var _ Workloader = (*K8S)(nil)
var _ Capaciter = (*K8S)(nil)

type NetworkResource struct {
	ReservationInfo `bson:",inline"`

	Name                         string            `bson:"name" json:"name"`
	StatsAggregator              []StatsAggregator `bson:"stats_aggregator" json:"stats_aggregator"`
	WireguardPrivateKeyEncrypted string            `bson:"wireguard_private_key_encrypted" json:"wireguard_private_key_encrypted"`
	WireguardPublicKey           string            `bson:"wireguard_public_key" json:"wireguard_public_key"`
	WireguardListenPort          int64             `bson:"wireguard_listen_port" json:"wireguard_listen_port"`
	Iprange                      schema.IPRange    `bson:"iprange" json:"iprange"`
	Peers                        []WireguardPeer   `bson:"peers" json:"peers"`
}

func (n *NetworkResource) GetRSU() RSU {
	return RSU{}
}

func (v *NetworkResource) VerifyJSON() error {
	dup := NetworkResource{}

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
