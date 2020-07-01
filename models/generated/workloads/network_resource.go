package workloads

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	schema "github.com/threefoldtech/tfexplorer/schema"
)

var _ Workloader = (*K8S)(nil)
var _ Capaciter = (*K8S)(nil)

type NetworkResource struct {
	ReservationInfo `bson:",inline"`

	Name                         string            `bson:"name" json:"name"`
	WireguardPrivateKeyEncrypted string            `bson:"wireguard_private_key_encrypted" json:"wireguard_private_key_encrypted"`
	WireguardPublicKey           string            `bson:"wireguard_public_key" json:"wireguard_public_key"`
	WireguardListenPort          int64             `bson:"wireguard_listen_port" json:"wireguard_listen_port"`
	Iprange                      schema.IPRange    `bson:"iprange" json:"iprange"`
	Peers                        []WireguardPeer   `bson:"peers" json:"peers"`
	StatsAggregator              []StatsAggregator `bson:"stats_aggregator" json:"stats_aggregator"`
}

func (n *NetworkResource) GetRSU() RSU {
	return RSU{}
}

func (n *NetworkResource) SignatureChallenge() ([]byte, error) {
	ric, err := n.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", n.Name); err != nil {
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
	for _, s := range n.StatsAggregator {
		if err := s.SigingEncode(b); err != nil {
			return nil, err
		}
	}

	h := sha256.New()
	if _, err := h.Write(b.Bytes()); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
