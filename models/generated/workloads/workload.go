package workloads

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
	schema "github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	Workloader interface {
		WorkloadID() int64
		GetWorkloadType() WorkloadTypeEnum
		GetID() schema.ID
		SetID(id schema.ID)
		GetJson() string
		SetJson(json string)
		GetCustomerTid() int64
		SetCustomerTid(tid int64)
		GetCustomerSignature() string
		SetCustomerSignature(signature string)
		GetNextAction() NextActionEnum
		SetNextAction(next NextActionEnum)
		GetSignaturesProvision() []SigningSignature
		PushSignatureProvision(signature SigningSignature)
		GetSignatureFarmer() SigningSignature
		SetSignatureFarmer(signature SigningSignature)
		GetSignaturesDelete() []SigningSignature
		PushSignatureDelete(signature SigningSignature)
		GetEpoch() schema.Date
		SetEpoch(date schema.Date)
		GetMetadata() string
		SetMetadata(metadata string)
		GetResult() Result
		SetResult(result Result)
		GetDescription() string
		SetDescription(description string)
		GetSigningRequestProvision() SigningRequest
		SetSigningRequestProvision(request SigningRequest)
		GetSigningRequestDelete() SigningRequest
		SetSigningRequestDelete(request SigningRequest)
		GetExpirationProvisioning() schema.Date
		SetExpirationProvisioning(date schema.Date)
		SetSignaturesProvision(signatures []SigningSignature)
		SetSignaturesDelete(signatuers []SigningSignature)
		VerifyJSON() error
		GetPoolID() int64
		GetNodeID() string

		Capaciter
	}

	Capaciter interface {
		GetRSU() RSU
	}

	RSU struct {
		CRU int64
		SRU int64
		HRU int64
		MRU float64
	}
)

// UnmarshalJSON decodes a workload from JSON format
func UnmarshalJSON(buffer []byte) (Workloader, error) {
	typeField := struct {
		WorkloadType WorkloadTypeEnum `json:"workload_type"`
	}{}

	if err := json.Unmarshal(buffer, &typeField); err != nil {
		return nil, errors.Wrap(err, "could not decode workload type")
	}

	var err error
	var workload Workloader

	switch typeField.WorkloadType {
	case WorkloadTypeContainer:
		var c Container
		err = json.Unmarshal(buffer, &c)
		workload = &c
	case WorkloadTypeDomainDelegate:
		var g GatewayDelegate
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeGateway4To6:
		var g Gateway4To6
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeKubernetes:
		var k K8S
		err = json.Unmarshal(buffer, &k)
		workload = &k
	case WorkloadTypeNetworkResource:
		var n NetworkResource
		err = json.Unmarshal(buffer, &n)
		workload = &n
	case WorkloadTypeProxy:
		var g GatewayProxy
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeReverseProxy:
		var g GatewayReverseProxy
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeSubDomain:
		var g GatewaySubdomain
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeVolume:
		var v Volume
		err = json.Unmarshal(buffer, &v)
		workload = &v
	case WorkloadTypeZDB:
		var z ZDB
		err = json.Unmarshal(buffer, &z)
		workload = &z
	default:
		return nil, errors.New("unrecognized workload type")
	}

	return workload, err
}

// UnmarshalBSON decodes a workload from BSON format
func UnmarshalBSON(buffer []byte) (Workloader, error) {
	typeField := struct {
		WorkloadType WorkloadTypeEnum `bson:"workload_type"`
	}{}

	if err := bson.Unmarshal(buffer, &typeField); err != nil {
		return nil, errors.Wrap(err, "could not decode workload type")
	}

	var err error
	var workload Workloader

	switch typeField.WorkloadType {
	case WorkloadTypeContainer:
		var c Container
		err = bson.Unmarshal(buffer, &c)
		workload = &c
	case WorkloadTypeDomainDelegate:
		var g GatewayDelegate
		err = bson.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeGateway4To6:
		var g Gateway4To6
		err = bson.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeKubernetes:
		var k K8S
		err = bson.Unmarshal(buffer, &k)
		workload = &k
	case WorkloadTypeNetworkResource:
		var n NetworkResource
		err = bson.Unmarshal(buffer, &n)
		workload = &n
	case WorkloadTypeProxy:
		var g GatewayProxy
		err = bson.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeReverseProxy:
		var g GatewayReverseProxy
		err = bson.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeSubDomain:
		var g GatewaySubdomain
		err = bson.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeVolume:
		var v Volume
		err = bson.Unmarshal(buffer, &v)
		workload = &v
	case WorkloadTypeZDB:
		var z ZDB
		err = bson.Unmarshal(buffer, &z)
		workload = &z
	default:
		return nil, errors.New("unrecognized workload type")
	}

	return workload, err
}

type ReservationInfo struct {
	WorkloadId int64  `bson:"workload_id" json:"workload_id"`
	NodeId     string `bson:"node_id" json:"node_id"`
	PoolId     int64  `bson:"pool_id" json:"pool_id"`

	Description             string         `bson:"description" json:"description"`
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

func (i *ReservationInfo) WorkloadID() int64 {
	return i.WorkloadId
}

func (i *ReservationInfo) GetWorkloadType() WorkloadTypeEnum {
	return i.WorkloadType
}

func (i *ReservationInfo) GetID() schema.ID {
	return i.ID
}

func (i *ReservationInfo) SetID(id schema.ID) {
	i.ID = id
}

func (i *ReservationInfo) GetJson() string {
	return i.Json
}

func (i *ReservationInfo) GetCustomerTid() int64 {
	return i.CustomerTid
}

func (i *ReservationInfo) GetCustomerSignature() string {
	return i.CustomerSignature
}

func (i *ReservationInfo) GetNextAction() NextActionEnum {
	return i.NextAction
}

func (i *ReservationInfo) SetNextAction(next NextActionEnum) {
	i.NextAction = next
}

func (i *ReservationInfo) GetSignaturesProvision() []SigningSignature {
	return i.SignaturesProvision
}

func (i *ReservationInfo) PushSignatureProvision(signature SigningSignature) {
	i.SignaturesProvision = append(i.SignaturesProvision, signature)
}

func (i *ReservationInfo) GetSignatureFarmer() SigningSignature {
	return i.SignatureFarmer
}

func (i *ReservationInfo) SetSignatureFarmer(signature SigningSignature) {
	i.SignatureFarmer = signature
}

func (i *ReservationInfo) GetSignaturesDelete() []SigningSignature {
	return i.SignaturesDelete
}

func (i *ReservationInfo) PushSignatureDelete(signature SigningSignature) {
	i.SignaturesDelete = append(i.SignaturesDelete, signature)
}

func (i *ReservationInfo) GetEpoch() schema.Date {
	return i.Epoch
}

func (i *ReservationInfo) GetMetadata() string {
	return i.Metadata
}

func (i *ReservationInfo) GetResult() Result {
	return i.Result
}

func (i *ReservationInfo) SetResult(result Result) {
	i.Result = result
}

func (i *ReservationInfo) GetDescription() string {
	return i.Description
}

func (i *ReservationInfo) GetSigningRequestProvision() SigningRequest {
	return i.SigningRequestProvision
}

func (i *ReservationInfo) GetSigningRequestDelete() SigningRequest {
	return i.SigningRequestDelete
}

func (i *ReservationInfo) GetExpirationProvisioning() schema.Date {
	return i.ExpirationProvisioning
}

func (i *ReservationInfo) SetJson(json string) {
	i.Json = json
}

func (i *ReservationInfo) SetCustomerTid(tid int64) {
	i.CustomerTid = tid
}

func (i *ReservationInfo) SetCustomerSignature(signature string) {
	i.CustomerSignature = signature
}

func (i *ReservationInfo) SetEpoch(date schema.Date) {
	i.Epoch = date
}

func (i *ReservationInfo) SetMetadata(metadata string) {
	i.Metadata = metadata
}

func (i *ReservationInfo) SetDescription(description string) {
	i.Description = description
}

func (i *ReservationInfo) SetSigningRequestProvision(request SigningRequest) {
	i.SigningRequestProvision = request
}

func (i *ReservationInfo) SetSigningRequestDelete(request SigningRequest) {
	i.SigningRequestDelete = request
}

func (i *ReservationInfo) SetExpirationProvisioning(date schema.Date) {
	i.ExpirationProvisioning = date
}

func (i *ReservationInfo) SetSignaturesProvision(signatures []SigningSignature) {
	i.SignaturesProvision = signatures
}

func (i *ReservationInfo) SetSignaturesDelete(signatures []SigningSignature) {
	i.SignaturesDelete = signatures
}

func (i *ReservationInfo) VerifyJSON() error {
	dup := Volume{}

	if err := json.Unmarshal([]byte(i.Json), &dup); err != nil {
		return errors.Wrap(err, "invalid json data")
	}

	// override the fields which are not part of the signature
	dup.ID = i.ID
	dup.Json = i.Json
	dup.CustomerTid = i.CustomerTid
	dup.NextAction = i.NextAction
	dup.SignaturesProvision = i.SignaturesProvision
	dup.SignatureFarmer = i.SignatureFarmer
	dup.SignaturesDelete = i.SignaturesDelete
	dup.Epoch = i.Epoch
	dup.Metadata = i.Metadata
	dup.Result = i.Result
	dup.WorkloadType = i.WorkloadType

	if match := reflect.DeepEqual(i, dup); !match {
		return errors.New("json data does not match actual data")
	}

	return nil
}

func (i *ReservationInfo) GetPoolID() int64 {
	return i.PoolId
}

func (i *ReservationInfo) GetNodeID() string {
	return i.NodeId
}
