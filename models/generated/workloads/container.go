package workloads

import (
	"encoding/json"
	"net"
	"reflect"

	"github.com/pkg/errors"
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

func (c *Container) WorkloadID() int64 {
	return c.WorkloadId
}

func (c *Container) GetWorkloadType() WorkloadTypeEnum {
	return c.WorkloadType
}

func (c *Container) GetID() schema.ID {
	return c.ID
}

func (c *Container) SetID(id schema.ID) {
	c.ID = id
}

func (c *Container) GetJson() string {
	return c.Json
}

func (c *Container) GetCustomerTid() int64 {
	return c.CustomerTid
}

func (c *Container) GetCustomerSignature() string {
	return c.CustomerSignature
}

func (c *Container) GetNextAction() NextActionEnum {
	return c.NextAction
}

func (c *Container) SetNextAction(next NextActionEnum) {
	c.NextAction = next
}

func (c *Container) GetSignaturesProvision() []SigningSignature {
	return c.SignaturesProvision
}

func (c *Container) PushSignatureProvision(signature SigningSignature) {
	c.SignaturesProvision = append(c.SignaturesProvision, signature)
}

func (c *Container) GetSignatureFarmer() SigningSignature {
	return c.SignatureFarmer
}

func (c *Container) SetSignatureFarmer(signature SigningSignature) {
	c.SignatureFarmer = signature
}

func (c *Container) GetSignaturesDelete() []SigningSignature {
	return c.SignaturesDelete
}

func (c *Container) PushSignatureDelete(signature SigningSignature) {
	c.SignaturesDelete = append(c.SignaturesDelete, signature)
}

func (c *Container) GetEpoch() schema.Date {
	return c.Epoch
}

func (c *Container) GetMetadata() string {
	return c.Metadata
}

func (c *Container) GetResult() Result {
	return c.Result
}

func (c *Container) SetResult(result Result) {
	c.Result = result
}

func (c *Container) GetDescription() string {
	return c.Description
}

func (c *Container) GetCurrencies() []string {
	return c.Currencies
}

func (c *Container) GetSigningRequestProvision() SigningRequest {
	return c.SigningRequestProvision
}

func (c *Container) GetSigningRequestDelete() SigningRequest {
	return c.SigningRequestDelete
}

func (c *Container) GetExpirationProvisioning() schema.Date {
	return c.ExpirationProvisioning
}

func (c *Container) SetJson(json string) {
	c.Json = json
}

func (c *Container) SetCustomerTid(tid int64) {
	c.CustomerTid = tid
}

func (c *Container) SetCustomerSignature(signature string) {
	c.CustomerSignature = signature
}

func (c *Container) SetEpoch(date schema.Date) {
	c.Epoch = date
}

func (c *Container) SetMetadata(metadata string) {
	c.Metadata = metadata
}

func (c *Container) SetDescription(description string) {
	c.Description = description
}

func (c *Container) SetCurrencies(currencies []string) {
	c.Currencies = currencies
}

func (c *Container) SetSigningRequestProvision(request SigningRequest) {
	c.SigningRequestProvision = request
}

func (c *Container) SetSigningRequestDelete(request SigningRequest) {
	c.SigningRequestDelete = request
}

func (c *Container) SetExpirationProvisioning(date schema.Date) {
	c.ExpirationProvisioning = date
}

func (c *Container) SetSignaturesProvision(signatures []SigningSignature) {
	c.SignaturesProvision = signatures
}

func (c *Container) SetSignaturesDelete(signatures []SigningSignature) {
	c.SignaturesDelete = signatures
}

func (c *Container) VerifyJSON() error {
	dup := Container{}

	if err := json.Unmarshal([]byte(c.Json), &dup); err != nil {
		return errors.Wrap(err, "invalid json data")
	}

	// override the fields which are not part of the signature
	dup.ID = c.ID
	dup.Json = c.Json
	dup.CustomerTid = c.CustomerTid
	dup.NextAction = c.NextAction
	dup.SignaturesProvision = c.SignaturesProvision
	dup.SignatureFarmer = c.SignatureFarmer
	dup.SignaturesDelete = c.SignaturesDelete
	dup.Epoch = c.Epoch
	dup.Metadata = c.Metadata
	dup.Result = c.Result
	dup.WorkloadType = c.WorkloadType

	if match := reflect.DeepEqual(c, dup); !match {
		return errors.New("json data does not match actual data")
	}

	return nil
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
