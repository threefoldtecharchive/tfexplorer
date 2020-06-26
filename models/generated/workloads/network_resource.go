package workloads

import schema "github.com/threefoldtech/tfexplorer/schema"

type NetworkResource struct {
	Name                         string            `bson:"name" json:"name"`
	WorkloadId                   int64             `bson:"workload_id" json:"workload_id"`
	NodeId                       string            `bson:"node_id" json:"node_id"`
	StatsAggregator              []StatsAggregator `bson:"stats_aggregator" json:"stats_aggregator"`
	WireguardPrivateKeyEncrypted string            `bson:"wireguard_private_key_encrypted" json:"wireguard_private_key_encrypted"`
	WireguardPublicKey           string            `bson:"wireguard_public_key" json:"wireguard_public_key"`
	WireguardListenPort          int64             `bson:"wireguard_listen_port" json:"wireguard_listen_port"`
	Iprange                      schema.IPRange    `bson:"iprange" json:"iprange"`
	Peers                        []WireguardPeer   `bson:"peers" json:"peers"`
	PoolId                       int64             `bson:"pool_id" json:"pool_id"`

	Description             string         `bson:"description" json:"description"`
	Currencies              []string       `bson:"currencies" json:"currencies"`
	SigningRequestProvision SigningRequest `bson:"signing_request_provision" json:"signing_request_provision"`
	SigningRequestDelete    SigningRequest `bson:"signing_request_delete" json:"signing_request_delete"`
	ExpirationProvisioning  schema.Date    `bson:"expiration_provisioning" json:"expiration_provisioning"`
	ID                      schema.ID      `bson:"_id" json:"id"`

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

func (n *NetworkResource) WorkloadID() int64 {
	return n.WorkloadId
}

func (n *NetworkResource) GetWorkloadType() WorkloadTypeEnum {
	return n.WorkloadType
}

func (n *NetworkResource) GetID() schema.ID {
	return n.ID
}

func (n *NetworkResource) SetID(id schema.ID) {
	n.ID = id
}

func (n *NetworkResource) GetJson() string {
	return n.Json
}

func (n *NetworkResource) GetCustomerTid() int64 {
	return n.CustomerTid
}

func (n *NetworkResource) GetCustomerSignature() string {
	return n.CustomerSignature
}

func (n *NetworkResource) GetNextAction() NextActionEnum {
	return n.NextAction
}

func (n *NetworkResource) SetNextAction(next NextActionEnum) {
	n.NextAction = next
}

func (n *NetworkResource) GetSignaturesProvision() []SigningSignature {
	return n.SignaturesProvision
}

func (n *NetworkResource) PushSignatureProvision(signature SigningSignature) {
	n.SignaturesProvision = append(n.SignaturesProvision, signature)
}

func (n *NetworkResource) GetSignatureFarmer() SigningSignature {
	return n.SignatureFarmer
}

func (n *NetworkResource) SetSignatureFarmer(signature SigningSignature) {
	n.SignatureFarmer = signature
}

func (n *NetworkResource) GetSignaturesDelete() []SigningSignature {
	return n.SignaturesDelete
}

func (n *NetworkResource) PushSignatureDelete(signature SigningSignature) {
	n.SignaturesDelete = append(n.SignaturesDelete, signature)
}

func (n *NetworkResource) GetEpoch() schema.Date {
	return n.Epoch
}

func (n *NetworkResource) GetMetadata() string {
	return n.Metadata
}

func (n *NetworkResource) GetResult() Result {
	return n.Result
}

func (n *NetworkResource) SetResult(result Result) {
	n.Result = result
}

func (n *NetworkResource) GetDescription() string {
	return n.Description
}

func (n *NetworkResource) GetCurrencies() []string {
	return n.Currencies
}

func (n *NetworkResource) GetSigningRequestProvision() SigningRequest {
	return n.SigningRequestProvision
}

func (n *NetworkResource) GetSigningRequestDelete() SigningRequest {
	return n.GetSigningRequestDelete()
}

func (n *NetworkResource) GetExpirationProvisioning() schema.Date {
	return n.ExpirationProvisioning
}
