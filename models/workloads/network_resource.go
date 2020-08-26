package workloads

import (
	"bytes"
	"fmt"

	schema "github.com/threefoldtech/tfexplorer/schema"
)

var _ Workloader = (*K8S)(nil)
var _ Capaciter = (*K8S)(nil)

type NetworkResource struct {
	contract Contract
	state    State

	Name                         string            `bson:"name" json:"name"`
	NetworkIprange               schema.IPRange    `bson:"network_iprange" json:"network_iprange"`
	WireguardPrivateKeyEncrypted string            `bson:"wireguard_private_key_encrypted" json:"wireguard_private_key_encrypted"`
	WireguardPublicKey           string            `bson:"wireguard_public_key" json:"wireguard_public_key"`
	WireguardListenPort          int64             `bson:"wireguard_listen_port" json:"wireguard_listen_port"`
	Iprange                      schema.IPRange    `bson:"iprange" json:"iprange"`
	Peers                        []WireguardPeer   `bson:"peers" json:"peers"`
	StatsAggregator              []StatsAggregator `bson:"stats_aggregator" json:"stats_aggregator"`
}

// Contract implements the Workloader interface
func (n *NetworkResource) Contract() *Contract { return &n.contract }

// State implements the Workloader interface
func (n *NetworkResource) State() *State { return &n.state }

// GetRSU implements the Capaciter interface
func (n *NetworkResource) GetRSU() RSU {
	return RSU{}
}

func (n *NetworkResource) SignatureChallenge() ([]byte, error) {
	ric, err := n.contract.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", n.Name); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", n.NetworkIprange.String()); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", n.WireguardPrivateKeyEncrypted); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", n.WireguardPublicKey); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%d", n.WireguardListenPort); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", n.Iprange.String()); err != nil {
		return nil, err
	}
	for _, p := range n.Peers {
		if err := p.SigingEncode(b); err != nil {
			return nil, err
		}
	}
	fmt.Println(string(b.String()))
	return b.Bytes(), nil
}
