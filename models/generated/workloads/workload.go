package workloads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

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
		SetSignaturesProvision(signatures []SigningSignature)
		SetSignaturesDelete(signatuers []SigningSignature)
		SignatureChallenge() ([]byte, error)
		SetPoolID(int64)
		GetPoolID() int64
		GetNodeID() string
		UniqueWorkloadID() string
		SetReference(string)
		GetReference() string

		Capaciter
	}

	Capaciter interface {
		GetRSU() RSU
	}

	RSU struct {
		CRU int64
		SRU float64
		HRU float64
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

	// Referene to an old reservation, used in conversion
	Reference string `bson:"reference" json:"reference"`

	Description             string         `bson:"description" json:"description"`
	SigningRequestProvision SigningRequest `bson:"signing_request_provision" json:"signing_request_provision"`
	SigningRequestDelete    SigningRequest `bson:"signing_request_delete" json:"signing_request_delete"`

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
	if request.Signers == nil {
		request.Signers = make([]int64, 0)
	}
	i.SigningRequestProvision = request
}

func (i *ReservationInfo) SetSigningRequestDelete(request SigningRequest) {
	if request.Signers == nil {
		request.Signers = make([]int64, 0)
	}
	i.SigningRequestDelete = request
}

func (i *ReservationInfo) SetSignaturesProvision(signatures []SigningSignature) {
	if signatures == nil {
		signatures = make([]SigningSignature, 0)
	}
	i.SignaturesProvision = signatures
}

func (i *ReservationInfo) SetSignaturesDelete(signatures []SigningSignature) {
	if signatures == nil {
		signatures = make([]SigningSignature, 0)
	}
	i.SignaturesDelete = signatures
}

// SignatureChallenge return a slice of byte containing all the date used to generate the
// signature of the workload
func (i *ReservationInfo) SignatureChallenge() ([]byte, error) { //TODO: name of this is shit
	b := &bytes.Buffer{}

	if _, err := fmt.Fprintf(b, "%d", i.WorkloadId); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", i.NodeId); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%d", i.PoolId); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", i.Reference); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%d", i.CustomerTid); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", strings.ToUpper(i.WorkloadType.String())); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%d", i.Epoch.Unix()); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", i.Description); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", i.Metadata); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (i *ReservationInfo) GetPoolID() int64 {
	return i.PoolId
}

func (i *ReservationInfo) GetNodeID() string {
	return i.NodeId
}

func (i *ReservationInfo) UniqueWorkloadID() string {
	return fmt.Sprintf("%d-%d", i.ID, i.WorkloadId)
}

func (i *ReservationInfo) SetPoolID(poolID int64) {
	i.PoolId = poolID
}

func (i *ReservationInfo) GetReference() string {
	return i.Reference
}

func (i *ReservationInfo) SetReference(ref string) {
	i.Reference = ref
}

// Stub type not used (for now)
type StatsAggregator struct {
	// To be defined
}

func (s StatsAggregator) SigingEncode(w io.Writer) error {
	return nil
}
