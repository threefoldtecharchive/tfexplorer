package workloads

import (
	"encoding/json"

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
		GetCurrencies() []string
		SetCurrencies(currencies []string)
		GetSigningRequestProvision() SigningRequest
		SetSigningRequestProvision(request SigningRequest)
		GetSigningRequestDelete() SigningRequest
		SetSigningRequestDelete(request SigningRequest)
		GetExpirationProvisioning() schema.Date
		SetExpirationProvisioning(date schema.Date)
		SetSignaturesProvision(signatures []SigningSignature)
		SetSignaturesDelete(signatuers []SigningSignature)
		VerifyJSON() error
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
