package workloads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	schema "github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	Workloader interface {
		State() *State
		Contract() *Contract
	}

	Capaciter interface {
		GetRSU() RSU
	}

	Signer interface {
		SignatureChallenge() ([]byte, error)
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

type Contract struct {
	ID                      schema.ID        `bson:"_id" json:"id"`
	WorkloadID              int64            `bson:"workload_id" json:"workload_id"`
	NodeID                  string           `bson:"node_id" json:"node_id"`
	PoolID                  int64            `bson:"pool_id" json:"pool_id"`
	SigningRequestProvision SigningRequest   `bson:"signing_request_provision" json:"signing_request_provision"`
	SigningRequestDelete    SigningRequest   `bson:"signing_request_delete" json:"signing_request_delete"`
	CustomerTid             int64            `bson:"customer_tid" json:"customer_tid"`
	WorkloadType            WorkloadTypeEnum `bson:"workload_type" json:"workload_type"`
	Epoch                   schema.Date      `bson:"epoch" json:"epoch"`
	Description             string           `bson:"description" json:"description"`
	Metadata                string           `bson:"metadata" json:"metadata"`

	// Referene to an old reservation, used in conversion
	Reference string `bson:"reference" json:"reference"`
}

func (c Contract) UniqueWorkloadID() string {
	return fmt.Sprintf("%d-%d", c.ID, c.WorkloadID)
}

type State struct {
	CustomerSignature   string             `bson:"customer_signature" json:"customer_signature"`
	NextAction          NextActionEnum     `bson:"next_action" json:"next_action"`
	SignaturesProvision []SigningSignature `bson:"signatures_provision" json:"signatures_provision"`
	SignatureFarmer     SigningSignature   `bson:"signature_farmer" json:"signature_farmer"`
	SignaturesDelete    []SigningSignature `bson:"signatures_delete" json:"signatures_delete"`
	Result              Result             `bson:"result" json:"result"`
}

func NewState() State {
	return State{
		NextAction:          NextActionCreate,
		SignaturesProvision: make([]SigningSignature, 0),
		SignaturesDelete:    make([]SigningSignature, 0),
	}
}

// SignatureChallenge return a slice of byte containing all the date used to generate the
// signature of the workload
func (c *Contract) SignatureChallenge() ([]byte, error) { //TODO: name of this is shit
	b := &bytes.Buffer{}

	if _, err := fmt.Fprintf(b, "%d", c.WorkloadID); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", c.NodeID); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%d", c.PoolID); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", c.Reference); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%d", c.CustomerTid); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", strings.ToUpper(c.WorkloadType.String())); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%d", c.Epoch.Unix()); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", c.Description); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", c.Metadata); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
