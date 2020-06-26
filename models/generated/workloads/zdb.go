package workloads

import schema "github.com/threefoldtech/tfexplorer/schema"

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

func (z *ZDB) WorkloadID() int64 {
	return z.WorkloadId
}

func (z *ZDB) GetWorkloadType() WorkloadTypeEnum {
	return z.WorkloadType
}

func (z *ZDB) GetID() schema.ID {
	return z.ID
}

func (z *ZDB) SetID(id schema.ID) {
	z.ID = id
}

func (z *ZDB) GetJson() string {
	return z.Json
}

func (z *ZDB) GetCustomerTid() int64 {
	return z.CustomerTid
}

func (z *ZDB) GetCustomerSignature() string {
	return z.CustomerSignature
}

func (z *ZDB) GetNextAction() NextActionEnum {
	return z.NextAction
}

func (z *ZDB) SetNextAction(next NextActionEnum) {
	z.NextAction = next
}

func (z *ZDB) GetSignaturesProvision() []SigningSignature {
	return z.SignaturesProvision
}

func (z *ZDB) PushSignatureProvision(signature SigningSignature) {
	z.SignaturesProvision = append(z.SignaturesProvision, signature)
}

func (z *ZDB) GetSignatureFarmer() SigningSignature {
	return z.SignatureFarmer
}

func (z *ZDB) SetSignatureFarmer(signature SigningSignature) {
	z.SignatureFarmer = signature
}

func (z *ZDB) GetSignaturesDelete() []SigningSignature {
	return z.SignaturesDelete
}

func (z *ZDB) PushSignatureDelete(signature SigningSignature) {
	z.SignaturesDelete = append(z.SignaturesDelete, signature)
}

func (z *ZDB) GetEpoch() schema.Date {
	return z.Epoch
}

func (z *ZDB) GetMetadata() string {
	return z.Metadata
}

func (z *ZDB) GetResult() Result {
	return z.Result
}

func (z *ZDB) SetResult(result Result) {
	z.Result = result
}

func (z *ZDB) GetDescription() string {
	return z.Description
}

func (z *ZDB) GetCurrencies() []string {
	return z.Currencies
}

func (z *ZDB) GetSigningRequestProvision() SigningRequest {
	return z.SigningRequestProvision
}

func (z *ZDB) GetSigningRequestDelete() SigningRequest {
	return z.GetSigningRequestDelete()
}

func (z *ZDB) GetExpirationProvisioning() schema.Date {
	return z.ExpirationProvisioning
}

func (z *ZDB) SetJson(json string) {
	z.Json = json
}

func (z *ZDB) SetCustomerTid(tid int64) {
	z.CustomerTid = tid
}

func (z *ZDB) SetCustomerSignature(signature string) {
	z.CustomerSignature = signature
}

func (z *ZDB) SetEpoch(date schema.Date) {
	z.Epoch = date
}

func (z *ZDB) SetMetadata(metadata string) {
	z.Metadata = metadata
}

func (z *ZDB) SetDescription(description string) {
	z.Description = description
}

func (z *ZDB) SetCurrencies(currencies []string) {
	z.Currencies = currencies
}

func (z *ZDB) SetSigningRequestProvision(request SigningRequest) {
	z.SigningRequestProvision = request
}

func (z *ZDB) SetSigningRequestDelete(request SigningRequest) {
	z.SigningRequestDelete = request
}

func (z *ZDB) SetExpirationProvisioning(date schema.Date) {
	z.ExpirationProvisioning = date
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
