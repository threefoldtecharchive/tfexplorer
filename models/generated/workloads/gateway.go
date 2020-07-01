package workloads

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

var _ Workloader = (*GatewayProxy)(nil)
var _ Capaciter = (*GatewayProxy)(nil)

type GatewayProxy struct {
	ReservationInfo `bson:",inline"`

	Domain  string `bson:"domain" json:"domain"`
	Addr    string `bson:"addr" json:"addr"`
	Port    uint32 `bson:"port" json:"port"`
	PortTLS uint32 `bson:"port_tls" json:"port_tls"`
}

func (g *GatewayProxy) GetRSU() RSU {
	return RSU{}
}

func (p *GatewayProxy) SignatureChallenge() ([]byte, error) {
	ric, err := p.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", p.Domain); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", p.Addr); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%d", p.Port); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%d", p.PortTLS); err != nil {
		return nil, err
	}

	h := sha256.New()
	if _, err := h.Write(b.Bytes()); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

var _ Workloader = (*GatewayReverseProxy)(nil)
var _ Capaciter = (*GatewayReverseProxy)(nil)

type GatewayReverseProxy struct {
	ReservationInfo `bson:",inline"`

	Domain string `bson:"domain" json:"domain"`
	Secret string `bson:"secret" json:"secret"`
}

func (g *GatewayReverseProxy) GetRSU() RSU {
	return RSU{}
}

func (p *GatewayReverseProxy) SignatureChallenge() ([]byte, error) {
	ric, err := p.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", p.Domain); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", p.Secret); err != nil {
		return nil, err
	}

	h := sha256.New()
	if _, err := h.Write(b.Bytes()); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

var _ Workloader = (*GatewaySubdomain)(nil)
var _ Capaciter = (*GatewaySubdomain)(nil)

type GatewaySubdomain struct {
	ReservationInfo `bson:",inline"`

	Domain string   `bson:"domain" json:"domain"`
	IPs    []string `bson:"ips" json:"ips"`
}

func (g *GatewaySubdomain) GetRSU() RSU {
	return RSU{}
}

func (s *GatewaySubdomain) SignatureChallenge() ([]byte, error) {
	ric, err := s.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", s.Domain); err != nil {
		return nil, err
	}
	for _, ip := range s.IPs {
		if _, err := fmt.Fprintf(b, "%s", ip); err != nil {
			return nil, err
		}
	}

	h := sha256.New()
	if _, err := h.Write(b.Bytes()); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

var _ Workloader = (*GatewayDelegate)(nil)
var _ Capaciter = (*GatewayDelegate)(nil)

type GatewayDelegate struct {
	ReservationInfo `bson:",inline"`

	Domain string `bson:"domain" json:"domain"`
}

func (g *GatewayDelegate) GetRSU() RSU {
	return RSU{}
}

func (d *GatewayDelegate) SignatureChallenge() ([]byte, error) {
	ric, err := d.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", d.Domain); err != nil {
		return nil, err
	}

	h := sha256.New()
	if _, err := h.Write(b.Bytes()); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

var _ Workloader = (*Gateway4To6)(nil)
var _ Capaciter = (*Gateway4To6)(nil)

type Gateway4To6 struct {
	ReservationInfo `bson:",inline"`

	PublicKey string `bson:"public_key" json:"public_key"`
}

func (g *Gateway4To6) GetRSU() RSU {
	return RSU{}
}

func (g *Gateway4To6) SignatureChallenge() ([]byte, error) {
	ric, err := g.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%s", g.PublicKey); err != nil {
		return nil, err
	}

	h := sha256.New()
	if _, err := h.Write(b.Bytes()); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
