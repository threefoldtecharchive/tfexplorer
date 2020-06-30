package workloads

var _ Workloader = (*GatewayProxy)(nil)
var _ Capaciter = (*GatewayProxy)(nil)

type GatewayProxy struct {
	ReservationInfo

	Domain  string `bson:"domain" json:"domain"`
	Addr    string `bson:"addr" json:"addr"`
	Port    uint32 `bson:"port" json:"port"`
	PortTLS uint32 `bson:"port_tls" json:"port_tls"`
	PoolId  int64  `bson:"pool_id" json:"pool_id"`
}

func (g *GatewayProxy) GetRSU() RSU {
	return RSU{}
}

var _ Workloader = (*GatewayReverseProxy)(nil)
var _ Capaciter = (*GatewayReverseProxy)(nil)

type GatewayReverseProxy struct {
	ReservationInfo

	Domain string `bson:"domain" json:"domain"`
	Secret string `bson:"secret" json:"secret"`
	PoolId int64  `bson:"pool_id" json:"pool_id"`
}

func (g *GatewayReverseProxy) GetRSU() RSU {
	return RSU{}
}

var _ Workloader = (*GatewaySubdomain)(nil)
var _ Capaciter = (*GatewaySubdomain)(nil)

type GatewaySubdomain struct {
	ReservationInfo

	Domain string   `bson:"domain" json:"domain"`
	IPs    []string `bson:"ips" json:"ips"`
	PoolId int64    `bson:"pool_id" json:"pool_id"`
}

func (g *GatewaySubdomain) GetRSU() RSU {
	return RSU{}
}

var _ Workloader = (*GatewayDelegate)(nil)
var _ Capaciter = (*GatewayDelegate)(nil)

type GatewayDelegate struct {
	ReservationInfo

	Domain string `bson:"domain" json:"domain"`
	PoolId int64  `bson:"pool_id" json:"pool_id"`
}

func (g *GatewayDelegate) GetRSU() RSU {
	return RSU{}
}

var _ Workloader = (*Gateway4To6)(nil)
var _ Capaciter = (*Gateway4To6)(nil)

type Gateway4To6 struct {
	ReservationInfo

	PublicKey string `bson:"public_key" json:"public_key"`
	PoolId    int64  `bson:"pool_id" json:"pool_id"`
}

func (g *Gateway4To6) GetRSU() RSU {
	return RSU{}
}
