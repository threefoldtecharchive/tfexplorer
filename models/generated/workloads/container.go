package workloads

import (
	"net"

	schema "github.com/threefoldtech/tfexplorer/schema"
)

type Container struct {
	WorkloadId        int64               `bson:"workload_id" json:"workload_id"`
	NodeId            string              `bson:"node_id" json:"node_id"`
	Flist             string              `bson:"flist" json:"flist"`
	HubUrl            string              `bson:"hub_url" json:"hub_url"`
	Environment       map[string]string   `bson:"environment" json:"environment"`
	SecretEnvironment map[string]string   `bson:"secret_environment" json:"secret_environment"`
	Entrypoint        string              `bson:"entrypoint" json:"entrypoint"`
	Interactive       bool                `bson:"interactive" json:"interactive"`
	Volumes           []ContainerMount    `bson:"volumes" json:"volumes"`
	NetworkConnection []NetworkConnection `bson:"network_connection" json:"network_connection"`
	StatsAggregator   []StatsAggregator   `bson:"stats_aggregator" json:"stats_aggregator"`
	Logs              []Logs              `bson:"logs" json:"logs"`
	FarmerTid         int64               `bson:"farmer_tid" json:"farmer_tid"`
	Capacity          ContainerCapacity   `bson:"capcity" json:"capacity"`
	PoolId            int64               `bson:"pool_id" json:"pool_id"`

	ID                  schema.ID          `bson:"_id" json:"id"`
	Json                string             `bson:"json" json:"json"`
	CustomerTid         int64              `bson:"customer_tid" json:"customer_tid"`
	CustomerSignature   string             `bson:"customer_signature" json:"customer_signature"`
	NextAction          NextActionEnum     `bson:"next_action" json:"next_action"`
	SignaturesProvision []SigningSignature `bson:"signatures_provision" json:"signatures_provision"`
	SignaturesFarmer    []SigningSignature `bson:"signatures_farmer" json:"signatures_farmer"`
	SignaturesDelete    []SigningSignature `bson:"signatures_delete" json:"signatures_delete"`
	Epoch               schema.Date        `bson:"epoch" json:"epoch"`
	Metadata            string             `bson:"metadata" json:"metadata"`
	Results             []Result           `bson:"results" json:"results"`
}

func (c Container) WorkloadID() int64 {
	return c.WorkloadId
}

type ContainerCapacity struct {
	Cpu      int64        `bson:"cpu" json:"cpu"`
	Memory   int64        `bson:"memory" json:"memory"`
	DiskSize uint64       `bson:"disk_size" json:"disk_size"`
	DiskType DiskTypeEnum `bson:"disk_type" json:"disk_type"`
}

type Logs struct {
	Type string    `bson:"type" json:"type"`
	Data LogsRedis `bson:"data" json:"data"`
}

type LogsRedis struct {
	Stdout string `bson:"stdout" json:"stdout"`
	Stderr string `bson:"stderr" json:"stderr"`
}

type ContainerMount struct {
	VolumeId   string `bson:"volume_id" json:"volume_id"`
	Mountpoint string `bson:"mountpoint" json:"mountpoint"`
}

type NetworkConnection struct {
	NetworkId string `bson:"network_id" json:"network_id"`
	Ipaddress net.IP `bson:"ipaddress" json:"ipaddress"`
	PublicIp6 bool   `bson:"public_ip6" json:"public_ip6"`
}

type StatsAggregator struct {
	Type string     `bson:"type" json:"type"`
	Data StatsRedis `bson:"data" json:"data"`
}

type StatsRedis struct {
	Endpoint string `bson:"stdout" json:"endpoint"`
}
