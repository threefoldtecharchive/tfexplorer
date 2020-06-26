package workloads

import (
	"net"

	schema "github.com/threefoldtech/tfexplorer/schema"
)

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
	PoolId          int64             `bson:"pool_id" json:"pool_id"`

	Description             string         `bson:"description" json:"description"`
	Currencies              []string       `bson:"currencies" json:"currencies"`
	SigningRequestProvision SigningRequest `bson:"signing_request_provision" json:"signing_request_provision"`
	SigningRequestDelete    SigningRequest `bson:"signing_request_delete" json:"signing_request_delete"`
	ExpirationProvisioning  schema.Date    `bson:"expiration_provisioning" json:"expiration_provisioning"`

	ID                  schema.ID          `bson:"_id" json:"id"`
	Json                string             `bson:"json" json:"json"`
	CustomerTid         int64              `bson:"customer_tid" json:"customer_tid"`
	CustomerSignature   string             `bson:"customer_signature" json:"customer_signature"`
	NextAction          NextActionEnum     `bson:"next_action" json:"next_action"`
	SignaturesProvision []SigningSignature `bson:"signatures_provision" json:"signatures_provision"`
	SignatureFarmer     SigningSignature   `bson:"signature_farmer" json:"signature_farmer"`
	SignaturesDelete    []SigningSignature `bson:"signatures_delete" json:"signatures_delete"`
	Epoch               schema.Date        `bson:"epoch" json:"epoch"`
	Metadata            string             `bson:"metadata" json:"metadata"`
	Result              Result             `bson:"result" json:"result"`
	WorkloadType        WorkloadTypeEnum   `bson:"workload_type" json:"workload_type"`
}

func (k *K8S) WorkloadID() int64 {
	return k.WorkloadId
}

func (k *K8S) GetWorkloadType() WorkloadTypeEnum {
	return k.WorkloadType
}

func (k *K8S) GetID() schema.ID {
	return k.ID
}

func (k *K8S) SetID(id schema.ID) {
	k.ID = id
}

func (k *K8S) GetJson() string {
	return k.Json
}

func (k *K8S) GetCustomerTid() int64 {
	return k.CustomerTid
}

func (k *K8S) GetCustomerSignature() string {
	return k.CustomerSignature
}

func (k *K8S) GetNextAction() NextActionEnum {
	return k.NextAction
}

func (k *K8S) SetNextAction(next NextActionEnum) {
	k.NextAction = next
}

func (k *K8S) GetSignaturesProvision() []SigningSignature {
	return k.SignaturesProvision
}

func (k *K8S) PushSignatureProvision(signature SigningSignature) {
	k.SignaturesProvision = append(k.SignaturesProvision, signature)
}

func (k *K8S) GetSignatureFarmer() SigningSignature {
	return k.SignatureFarmer
}

func (k *K8S) SetSignatureFarmer(signature SigningSignature) {
	k.SignatureFarmer = signature
}

func (k *K8S) GetSignaturesDelete() []SigningSignature {
	return k.SignaturesDelete
}

func (k *K8S) PushSignatureDelete(signature SigningSignature) {
	k.SignaturesDelete = append(k.SignaturesDelete, signature)
}

func (k *K8S) GetEpoch() schema.Date {
	return k.Epoch
}

func (k *K8S) GetMetadata() string {
	return k.Metadata
}

func (k *K8S) GetResult() Result {
	return k.Result
}

func (k *K8S) SetResult(result Result) {
	k.Result = result
}

func (k *K8S) GetDescription() string {
	return k.Description
}

func (k *K8S) GetCurrencies() []string {
	return k.Currencies
}

func (k *K8S) GetSigningRequestProvision() SigningRequest {
	return k.SigningRequestProvision
}

func (k *K8S) GetSigningRequestDelete() SigningRequest {
	return k.SigningRequestDelete
}

func (k *K8S) GetExpirationProvisioning() schema.Date {
	return k.ExpirationProvisioning
}

func (k *K8S) SetJson(json string) {
	k.Json = json
}

func (k *K8S) SetCustomerTid(tid int64) {
	k.CustomerTid = tid
}

func (k *K8S) SetCustomerSignature(signature string) {
	k.CustomerSignature = signature
}

func (k *K8S) SetEpoch(date schema.Date) {
	k.Epoch = date
}

func (k *K8S) SetMetadata(metadata string) {
	k.Metadata = metadata
}

func (k *K8S) SetDescription(description string) {
	k.Description = description
}

func (k *K8S) SetCurrencies(currencies []string) {
	k.Currencies = currencies
}

func (k *K8S) SetSigningRequestProvision(request SigningRequest) {
	k.SigningRequestProvision = request
}

func (k *K8S) SetSigningRequestDelete(request SigningRequest) {
	k.SigningRequestDelete = request
}

func (k *K8S) SetExpirationProvisioning(date schema.Date) {
	k.ExpirationProvisioning = date
}
