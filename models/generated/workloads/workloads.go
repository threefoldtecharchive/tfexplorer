package workloads

import (
	"encoding/json"
	"net"

	schema "github.com/threefoldtech/tfexplorer/schema"
)

type Reservation struct {
	ID                  schema.ID          `bson:"_id" json:"id"`
	Json                string             `bson:"json" json:"json"`
	DataReservation     ReservationData    `bson:"data_reservation" json:"data_reservation"`
	CustomerTid         int64              `bson:"customer_tid" json:"customer_tid"`
	CustomerSignature   string             `bson:"customer_signature" json:"customer_signature"`
	NextAction          NextActionEnum     `bson:"next_action" json:"next_action"`
	SignaturesProvision []SigningSignature `bson:"signatures_provision" json:"signatures_provision"`
	SignaturesFarmer    []SigningSignature `bson:"signatures_farmer" json:"signatures_farmer"`
	SignaturesDelete    []SigningSignature `bson:"signatures_delete" json:"signatures_delete"`
	Epoch               schema.Date        `bson:"epoch" json:"epoch"`
	Results             []Result           `bson:"results" json:"results"`
}

type ReservationData struct {
	Description             string                `bson:"description" json:"description"`
	Currencies              []string              `bson:"currencies" json:"currencies"`
	SigningRequestProvision SigningRequest        `bson:"signing_request_provision" json:"signing_request_provision"`
	SigningRequestDelete    SigningRequest        `bson:"signing_request_delete" json:"signing_request_delete"`
	Containers              []Container           `bson:"containers" json:"containers"`
	Volumes                 []Volume              `bson:"volumes" json:"volumes"`
	Zdbs                    []ZDB                 `bson:"zdbs" json:"zdbs"`
	Networks                []Network             `bson:"networks" json:"networks"`
	Kubernetes              []K8S                 `bson:"kubernetes" json:"kubernetes"`
	Proxies                 []GatewayProxy        `bson:"proxies" json:"proxies"`
	ReserveProxy            []GatewayReserveProxy `bson:"reserve_proxies" json:"reserve_proxies"`
	Subdomains              []GatewaySubdomain    `bson:"subdomains" json:"subdomains"`
	DomainDelegates         []GatewayDelegate     `bson:"domain_delegates" json:"domain_delegates"`
	ExpirationProvisioning  schema.Date           `bson:"expiration_provisioning" json:"expiration_provisioning"`
	ExpirationReservation   schema.Date           `bson:"expiration_reservation" json:"expiration_reservation"`
}

type SigningRequest struct {
	Signers   []int64 `bson:"signers" json:"signers"`
	QuorumMin int64   `bson:"quorum_min" json:"quorum_min"`
}

type SigningSignature struct {
	Tid       int64       `bson:"tid" json:"tid"`
	Signature string      `bson:"signature" json:"signature"`
	Epoch     schema.Date `bson:"epoch" json:"epoch"`
}

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
}

func (c Container) WorkloadID() int64 {
	return c.WorkloadId
}

type ContainerCapacity struct {
	Cpu    int64 `bson:"cpu" json:"cpu"`
	Memory int64 `bson:"memory" json:"memory"`
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

type K8S struct {
	WorkloadId      int64             `bson:"workload_id" json:"workload_id"`
	NodeId          string            `bson:"node_id" json:"node_id"`
	Size            int64             `bson:"size" json:"size"`
	NetworkId       string            `bson:"network_id" json:"network_id"`
	Ipaddress       net.IP            `bson:"ipaddress" json:"ipaddress"`
	ClusterSecret   string            `bson:"cluster_secret" json:"cluster_secret"`
	MasterIps       []net.IP          `bson:"master_ips" json:"master_ips"`
	SshKeys         []string          `bson:"ssh_keys" json:"ssh_keys"`
	StatsAggregator []StatsAggregator `bson:"stats_aggregator" json:"stats_aggregator"`
	FarmerTid       int64             `bson:"farmer_tid" json:"farmer_tid"`
}

func (k K8S) WorkloadID() int64 {
	return k.WorkloadId
}

type Network struct {
	Name             string               `bson:"name" json:"name"`
	WorkloadId       int64                `bson:"workload_id" json:"workload_id"`
	Iprange          schema.IPRange       `bson:"iprange" json:"iprange"`
	StatsAggregator  []StatsAggregator    `bson:"stats_aggregator" json:"stats_aggregator"`
	NetworkResources []NetworkNetResource `bson:"network_resources" json:"network_resources"`
	FarmerTid        int64                `bson:"farmer_tid" json:"farmer_tid"`
}

func (n Network) WorkloadID() int64 {
	return n.WorkloadId
}

type NetworkNetResource struct {
	NodeId                       string          `bson:"node_id" json:"node_id"`
	WireguardPrivateKeyEncrypted string          `bson:"wireguard_private_key_encrypted" json:"wireguard_private_key_encrypted"`
	WireguardPublicKey           string          `bson:"wireguard_public_key" json:"wireguard_public_key"`
	WireguardListenPort          int64           `bson:"wireguard_listen_port" json:"wireguard_listen_port"`
	Iprange                      schema.IPRange  `bson:"iprange" json:"iprange"`
	Peers                        []WireguardPeer `bson:"peers" json:"peers"`
}

type WireguardPeer struct {
	PublicKey      string           `bson:"public_key" json:"public_key"`
	AllowedIprange []schema.IPRange `bson:"allowed_iprange" json:"allowed_iprange"`
	Endpoint       string           `bson:"endpoint" json:"endpoint"`
	Iprange        schema.IPRange   `bson:"iprange" json:"iprange"`
}

type Result struct {
	Category   ResultCategoryEnum `bson:"category" json:"category"`
	WorkloadId string             `bson:"workload_id" json:"workload_id"`
	DataJson   json.RawMessage    `bson:"data_json" json:"data_json"`
	Signature  string             `bson:"signature" json:"signature"`
	State      ResultStateEnum    `bson:"state" json:"state"`
	Message    string             `bson:"message" json:"message"`
	Epoch      schema.Date        `bson:"epoch" json:"epoch"`
	NodeId     string             `bson:"node_id" json:"node_id"`
}

type StatsAggregator struct {
	Type string     `bson:"type" json:"type"`
	Data StatsRedis `bson:"data" json:"data"`
}

type StatsRedis struct {
	Endpoint string `bson:"stdout" json:"endpoint"`
}

type Volume struct {
	WorkloadId      int64             `bson:"workload_id" json:"workload_id"`
	NodeId          string            `bson:"node_id" json:"node_id"`
	Size            int64             `bson:"size" json:"size"`
	Type            VolumeTypeEnum    `bson:"type" json:"type"`
	StatsAggregator []StatsAggregator `bson:"stats_aggregator" json:"stats_aggregator"`
	FarmerTid       int64             `bson:"farmer_tid" json:"farmer_tid"`
}

func (v Volume) WorkloadID() int64 {
	return v.WorkloadId
}

// NOTE: this type has some manual changes
// that need to be preserved between regenerations.
type ReservationWorkload struct {
	WorkloadId string           `bson:"workload_id" json:"workload_id"`
	User       string           `bson:"user" json:"user"`
	Type       WorkloadTypeEnum `bson:"type" json:"type"`
	Content    interface{}      `bson:"content" json:"content"`
	Created    schema.Date      `bson:"created" json:"created"`
	Duration   int64            `bson:"duration" json:"duration"`
	Signature  string           `bson:"signature" json:"signature"`
	ToDelete   bool             `bson:"to_delete" json:"to_delete"`
}

type ZDB struct {
	WorkloadId      int64             `bson:"workload_id" json:"workload_id"`
	NodeId          string            `bson:"node_id" json:"node_id"`
	Size            int64             `bson:"size" json:"size"`
	Mode            ZDBModeEnum       `bson:"mode" json:"mode"`
	Password        string            `bson:"password" json:"password"`
	DiskType        DiskTypeEnum      `bson:"disk_type" json:"disk_type"`
	Public          bool              `bson:"public" json:"public"`
	StatsAggregator []StatsAggregator `bson:"stats_aggregator" json:"stats_aggregator"`
	FarmerTid       int64             `bson:"farmer_tid" json:"farmer_tid"`
}

func (z ZDB) WorkloadID() int64 {
	return z.WorkloadId
}

type DiskTypeEnum uint8

const (
	DiskTypeHDD DiskTypeEnum = iota
	DiskTypeSSD
)

func (e DiskTypeEnum) String() string {
	switch e {
	case DiskTypeHDD:
		return "hdd"
	case DiskTypeSSD:
		return "ssd"
	}
	return "UNKNOWN"
}

type NextActionEnum uint8

const (
	NextActionCreate NextActionEnum = iota
	NextActionSign
	NextActionPay
	NextActionDeploy
	NextActionDelete
	NextActionInvalid
	NextActionDeleted
)

func (e NextActionEnum) String() string {
	switch e {
	case NextActionCreate:
		return "create"
	case NextActionSign:
		return "sign"
	case NextActionPay:
		return "pay"
	case NextActionDeploy:
		return "deploy"
	case NextActionDelete:
		return "delete"
	case NextActionInvalid:
		return "invalid"
	case NextActionDeleted:
		return "deleted"
	}
	return "UNKNOWN"
}

type ResultCategoryEnum uint8

const (
	ResultCategoryZDB ResultCategoryEnum = iota
	ResultCategoryContainer
	ResultCategoryNetwork
	ResultCategoryVolume
	ResultCategoryK8S
	ResultCategoryProxy
	ResultCategoryReverseProxy
	ResultCategorySubDomain
	ResultCategoryDomainDelegate
)

func (e ResultCategoryEnum) String() string {
	switch e {
	case ResultCategoryZDB:
		return "zdb"
	case ResultCategoryContainer:
		return "container"
	case ResultCategoryNetwork:
		return "network"
	case ResultCategoryVolume:
		return "volume"
	case ResultCategoryK8S:
		return "kubernetes"
	case ResultCategoryProxy:
		return "proxy"
	case ResultCategoryReverseProxy:
		return "reverse-proxy"
	case ResultCategorySubDomain:
		return "subdomain"
	case ResultCategoryDomainDelegate:
		return "domain-delegate"
	}
	return "UNKNOWN"
}

type ResultStateEnum uint8

const (
	ResultStateError ResultStateEnum = iota
	ResultStateOK
	ResultStateDeleted
)

func (e ResultStateEnum) String() string {
	switch e {
	case ResultStateError:
		return "error"
	case ResultStateOK:
		return "ok"
	case ResultStateDeleted:
		return "deleted"
	}
	return "UNKNOWN"
}

type VolumeTypeEnum uint8

const (
	VolumeTypeHDD VolumeTypeEnum = iota
	VolumeTypeSSD
)

func (e VolumeTypeEnum) String() string {
	switch e {
	case VolumeTypeHDD:
		return "HDD"
	case VolumeTypeSSD:
		return "SSD"
	}
	return "UNKNOWN"
}

type WorkloadTypeEnum uint8

const (
	WorkloadTypeZDB WorkloadTypeEnum = iota
	WorkloadTypeContainer
	WorkloadTypeVolume
	WorkloadTypeNetwork
	WorkloadTypeKubernetes
	WorkloadTypeProxy
	WorkloadTypeReverseProxy
	WorkloadTypeSubDomain
	WorkloadTypeDomainDelegate
)

func (e WorkloadTypeEnum) String() string {
	switch e {
	case WorkloadTypeZDB:
		return "zdb"
	case WorkloadTypeContainer:
		return "container"
	case WorkloadTypeVolume:
		return "volume"
	case WorkloadTypeNetwork:
		return "network"
	case WorkloadTypeKubernetes:
		return "kubernetes"
	case WorkloadTypeProxy:
		return "proxy"
	case WorkloadTypeReverseProxy:
		return "reverse-proxy"
	case WorkloadTypeSubDomain:
		return "subdomain"
	case WorkloadTypeDomainDelegate:
		return "domain-delegate"

	}
	return "UNKNOWN"
}

type ZDBModeEnum uint8

const (
	ZDBModeSeq ZDBModeEnum = iota
	ZDBModeUser
)

func (e ZDBModeEnum) String() string {
	switch e {
	case ZDBModeSeq:
		return "seq"
	case ZDBModeUser:
		return "user"
	}
	return "UNKNOWN"
}

type GatewayProxy struct {
	ID         schema.ID `bson:"_id" json:"id"`
	WorkloadId int64     `bson:"workload_id" json:"workload_id"`
	NodeId     string    `bson:"node_id" json:"node_id"`
	Domain     string    `bson:"domain" json:"domain"`
	Addr       string    `bson:"addr" json:"addr"`
	Port       uint32    `bson:"port" json:"port"`
	PortTLS    uint32    `bson:"port_tls" json:"port_tls"`
}

func (g GatewayProxy) WorkloadID() int64 {
	return g.WorkloadId
}

type GatewayReserveProxy struct {
	ID         schema.ID `bson:"_id" json:"id"`
	WorkloadId int64     `bson:"workload_id" json:"workload_id"`
	NodeId     string    `bson:"node_id" json:"node_id"`
	Domain     string    `bson:"domain" json:"domain"`
	Secret     string    `bson:"secret" json:"secret"`
}

func (g GatewayReserveProxy) WorkloadID() int64 {
	return g.WorkloadId
}

type GatewaySubdomain struct {
	ID         schema.ID `bson:"_id" json:"id"`
	WorkloadId int64     `bson:"workload_id" json:"workload_id"`
	NodeId     string    `bson:"node_id" json:"node_id"`
	Domain     string    `bson:"domain" json:"domain"`
	IPs        []string  `bson:"ips" json:"ips"`
}

func (g GatewaySubdomain) WorkloadID() int64 {
	return g.WorkloadId
}

type GatewayDelegate struct {
	ID         schema.ID `bson:"_id" json:"id"`
	WorkloadId int64     `bson:"workload_id" json:"workload_id"`
	NodeId     string    `bson:"node_id" json:"node_id"`
	Domain     string    `bson:"domain" json:"domain"`
}

func (g GatewayDelegate) WorkloadID() int64 {
	return g.WorkloadId
}
