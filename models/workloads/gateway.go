package workloads

import (
	"bytes"
	"fmt"
)

var _ Workloader = (*GatewayProxy)(nil)
var _ Capaciter = (*GatewayProxy)(nil)

type GatewayProxy struct {
	contract Contract
	state    State

	Domain  string `bson:"domain" json:"domain"`
	Addr    string `bson:"addr" json:"addr"`
	Port    uint32 `bson:"port" json:"port"`
	PortTLS uint32 `bson:"port_tls" json:"port_tls"`
}

// Contract implements the Workloader interface
func (g *GatewayProxy) Contract() *Contract { return &g.contract }

// State implements the Workloader interface
func (g *GatewayProxy) State() *State { return &g.state }

// GetRSU implements the Capaciter interface
func (g *GatewayProxy) GetRSU() RSU {
	return RSU{}
}

func (g *GatewayProxy) SignatureChallenge() ([]byte, error) {
	ric, err := g.contract.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", g.Domain); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", g.Addr); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%d", g.Port); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%d", g.PortTLS); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

var _ Workloader = (*GatewayReverseProxy)(nil)
var _ Capaciter = (*GatewayReverseProxy)(nil)

type GatewayReverseProxy struct {
	contract Contract
	state    State

	Domain string `bson:"domain" json:"domain"`
	Secret string `bson:"secret" json:"secret"`
}

// Contract implements the Workloader interface
func (g *GatewayReverseProxy) Contract() *Contract { return &g.contract }

// State implements the Workloader interface
func (g *GatewayReverseProxy) State() *State { return &g.state }

// GetRSU implements the Capaciter interface
func (g *GatewayReverseProxy) GetRSU() RSU {
	return RSU{}
}

func (g *GatewayReverseProxy) SignatureChallenge() ([]byte, error) {
	ric, err := g.contract.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", g.Domain); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", g.Secret); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

var _ Workloader = (*GatewaySubdomain)(nil)
var _ Capaciter = (*GatewaySubdomain)(nil)

type GatewaySubdomain struct {
	contract Contract
	state    State

	Domain string   `bson:"domain" json:"domain"`
	IPs    []string `bson:"ips" json:"ips"`
}

// Contract implements the Workloader interface
func (g *GatewaySubdomain) Contract() *Contract { return &g.contract }

// State implements the Workloader interface
func (g *GatewaySubdomain) State() *State { return &g.state }

// GetRSU implements the Capaciter interface
func (g *GatewaySubdomain) GetRSU() RSU {
	return RSU{}
}

func (g *GatewaySubdomain) SignatureChallenge() ([]byte, error) {
	ric, err := g.contract.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", g.Domain); err != nil {
		return nil, err
	}
	for _, ip := range g.IPs {
		if _, err := fmt.Fprintf(b, "%s", ip); err != nil {
			return nil, err
		}
	}

	return b.Bytes(), nil
}

var _ Workloader = (*GatewayDelegate)(nil)
var _ Capaciter = (*GatewayDelegate)(nil)

type GatewayDelegate struct {
	contract Contract
	state    State

	Domain string `bson:"domain" json:"domain"`
}

// Contract implements the Workloader interface
func (g *GatewayDelegate) Contract() *Contract { return &g.contract }

// State implements the Workloader interface
func (g *GatewayDelegate) State() *State { return &g.state }

// GetRSU implements the Capaciter interface
func (g *GatewayDelegate) GetRSU() RSU {
	return RSU{}
}

func (d *GatewayDelegate) SignatureChallenge() ([]byte, error) {
	ric, err := d.contract.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", d.Domain); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

var _ Workloader = (*Gateway4To6)(nil)
var _ Capaciter = (*Gateway4To6)(nil)

type Gateway4To6 struct {
	contract Contract
	state    State

	PublicKey string `bson:"public_key" json:"public_key"`
}

// Contract implements the Workloader interface
func (g *Gateway4To6) Contract() *Contract { return &g.contract }

// State implements the Workloader interface
func (g *Gateway4To6) State() *State { return &g.state }

// GetRSU implements the Capaciter interface
func (g *Gateway4To6) GetRSU() RSU {
	return RSU{}
}

func (g *Gateway4To6) SignatureChallenge() ([]byte, error) {
	ric, err := g.contract.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", g.PublicKey); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
