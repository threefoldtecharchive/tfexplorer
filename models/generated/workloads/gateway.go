package workloads

import schema "github.com/threefoldtech/tfexplorer/schema"

type GatewayProxy struct {
	ID         schema.ID `bson:"_id" json:"id"`
	WorkloadId int64     `bson:"workload_id" json:"workload_id"`
	NodeId     string    `bson:"node_id" json:"node_id"`
	Domain     string    `bson:"domain" json:"domain"`
	Addr       string    `bson:"addr" json:"addr"`
	Port       uint32    `bson:"port" json:"port"`
	PortTLS    uint32    `bson:"port_tls" json:"port_tls"`
	PoolId     int64     `bson:"pool_id" json:"pool_id"`
}

func (g GatewayProxy) WorkloadID() int64 {
	return g.WorkloadId
}

type GatewayReverseProxy struct {
	ID         schema.ID `bson:"_id" json:"id"`
	WorkloadId int64     `bson:"workload_id" json:"workload_id"`
	NodeId     string    `bson:"node_id" json:"node_id"`
	Domain     string    `bson:"domain" json:"domain"`
	Secret     string    `bson:"secret" json:"secret"`
	PoolId     int64     `bson:"pool_id" json:"pool_id"`
}

func (g GatewayReverseProxy) WorkloadID() int64 {
	return g.WorkloadId
}

type GatewaySubdomain struct {
	ID         schema.ID `bson:"_id" json:"id"`
	WorkloadId int64     `bson:"workload_id" json:"workload_id"`
	NodeId     string    `bson:"node_id" json:"node_id"`
	Domain     string    `bson:"domain" json:"domain"`
	IPs        []string  `bson:"ips" json:"ips"`
	PoolId     int64     `bson:"pool_id" json:"pool_id"`
}

func (g GatewaySubdomain) WorkloadID() int64 {
	return g.WorkloadId
}

type GatewayDelegate struct {
	ID         schema.ID `bson:"_id" json:"id"`
	WorkloadId int64     `bson:"workload_id" json:"workload_id"`
	NodeId     string    `bson:"node_id" json:"node_id"`
	Domain     string    `bson:"domain" json:"domain"`
	PoolId     int64     `bson:"pool_id" json:"pool_id"`
}

func (g GatewayDelegate) WorkloadID() int64 {
	return g.WorkloadId
}

type Gateway4To6 struct {
	ID         schema.ID `bson:"_id" json:"id"`
	WorkloadId int64     `bson:"workload_id" json:"workload_id"`
	NodeId     string    `bson:"node_id" json:"node_id"`
	PublicKey  string    `bson:"public_key" json:"public_key"`
	PoolId     int64     `bson:"pool_id" json:"pool_id"`
}

func (g Gateway4To6) WorkloadID() int64 {
	return g.WorkloadId
}
