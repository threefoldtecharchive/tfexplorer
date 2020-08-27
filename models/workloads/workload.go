package workloads

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/threefoldtech/zos/pkg/crypto"

	"github.com/pkg/errors"

	schema "github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	Workloader interface {
		State() *State
		Contract() *Contract

		// All workload should be able to generate a challenge used for signature
		SignatureChallenge() ([]byte, error)
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
	base := struct {
		Contract
		State
	}{}

	if err := json.Unmarshal(buffer, &base); err != nil {
		return nil, errors.Wrap(err, "could not decode workload type")
	}

	var err error
	var workload Workloader

	switch base.WorkloadType {
	case WorkloadTypeContainer:
		var c Container
		c.state = base.State
		c.contract = base.Contract
		err = json.Unmarshal(buffer, &c)
		workload = &c
	case WorkloadTypeDomainDelegate:
		var g GatewayDelegate
		g.state = base.State
		g.contract = base.Contract
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeGateway4To6:
		var g Gateway4To6
		g.state = base.State
		g.contract = base.Contract
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeKubernetes:
		var k K8S
		k.state = base.State
		k.contract = base.Contract
		err = json.Unmarshal(buffer, &k)
		workload = &k
	case WorkloadTypeNetworkResource:
		var n NetworkResource
		n.state = base.State
		n.contract = base.Contract
		err = json.Unmarshal(buffer, &n)
		workload = &n
	case WorkloadTypeProxy:
		var g GatewayProxy
		g.state = base.State
		g.contract = base.Contract
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeReverseProxy:
		var g GatewayReverseProxy
		g.state = base.State
		g.contract = base.Contract
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeSubDomain:
		var g GatewaySubdomain
		g.state = base.State
		g.contract = base.Contract
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeVolume:
		var v Volume
		v.state = base.State
		v.contract = base.Contract
		err = json.Unmarshal(buffer, &v)
		workload = &v
	case WorkloadTypeZDB:
		var z ZDB
		z.state = base.State
		z.contract = base.Contract
		err = json.Unmarshal(buffer, &z)
		workload = &z
	default:
		return nil, errors.New("unrecognized workload type")
	}

	return workload, err
}

// UnmarshalBSON decodes a workload from BSON format
func UnmarshalBSON(buffer []byte) (Workloader, error) {
	base := struct {
		Contract `bson:",inline"`
		State    `bson:",inline"`
	}{}

	fmt.Printf("%s\n", string(buffer))
	if err := bson.Unmarshal(buffer, &base); err != nil {
		return nil, errors.Wrap(err, "could not decode workload type")
	}

	var err error
	var workload Workloader

	switch base.WorkloadType {
	case WorkloadTypeContainer:
		var c Container
		c.contract = base.Contract
		c.state = base.State
		err = bson.Unmarshal(buffer, &c)
		workload = &c
	case WorkloadTypeDomainDelegate:
		var g GatewayDelegate
		g.contract = base.Contract
		g.state = base.State
		err = bson.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeGateway4To6:
		var g Gateway4To6
		g.contract = base.Contract
		g.state = base.State
		err = bson.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeKubernetes:
		var k K8S
		k.contract = base.Contract
		k.state = base.State
		err = bson.Unmarshal(buffer, &k)
		workload = &k
	case WorkloadTypeNetworkResource:
		var n NetworkResource
		n.contract = base.Contract
		n.state = base.State
		err = bson.Unmarshal(buffer, &n)
		workload = &n
	case WorkloadTypeProxy:
		var g GatewayProxy
		g.contract = base.Contract
		g.state = base.State
		err = bson.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeReverseProxy:
		var g GatewayReverseProxy
		g.contract = base.Contract
		g.state = base.State
		err = bson.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeSubDomain:
		var g GatewaySubdomain
		g.contract = base.Contract
		g.state = base.State
		err = bson.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeVolume:
		var v Volume
		v.contract = base.Contract
		v.state = base.State
		err = bson.Unmarshal(buffer, &v)
		workload = &v
	case WorkloadTypeZDB:
		var z ZDB
		z.contract = base.Contract
		z.state = base.State
		err = bson.Unmarshal(buffer, &z)
		workload = &z
	default:
		return nil, errors.New("unrecognized workload type")
	}

	return workload, err
}

// Contract is the immutable part of an IT contract
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

// State is the mutable part of an IT contract
type State struct {
	CustomerSignature   string             `bson:"customer_signature" json:"customer_signature"`
	NextAction          NextActionEnum     `bson:"next_action" json:"next_action"`
	SignaturesProvision []SigningSignature `bson:"signatures_provision" json:"signatures_provision"`
	SignatureFarmer     SigningSignature   `bson:"signature_farmer" json:"signature_farmer"`
	SignaturesDelete    []SigningSignature `bson:"signatures_delete" json:"signatures_delete"`
	Result              Result             `bson:"result" json:"result"`
}

// NewState create a State object with proper default values
func NewState() State {
	return State{
		NextAction:          NextActionCreate,
		SignaturesProvision: make([]SigningSignature, 0),
		SignaturesDelete:    make([]SigningSignature, 0),
	}
}

// IsAny checks if the state NextAction value is any of the given status
func (s *State) IsAny(status ...NextActionEnum) bool {
	for _, x := range status {
		if s.NextAction == x {
			return true
		}
	}

	return false
}

// SignatureProvisionRequestVerify verify the signature from a signature request
// this is used for provision
// the signature is created from the workload siging challenge + "provision" + customer tid
func SignatureProvisionRequestVerify(w Workloader, pk string, sig SigningSignature) error {
	key, err := crypto.KeyFromHex(pk)
	if err != nil {
		return errors.Wrap(err, "invalid verification key")
	}

	b, err := w.SignatureChallenge()
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	if _, err := buf.WriteString("provision"); err != nil {
		return err
	}
	if _, err := buf.WriteString(fmt.Sprintf("%d", sig.Tid)); err != nil {
		return err
	}

	msg := sha256.Sum256(buf.Bytes())
	signature, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return err
	}

	return crypto.Verify(key, msg[:], signature)
}

// SignatureDeleteRequestVerify verify the signature from a signature request
// this is used for workload delete
// the signature is created from the workload siging challenge + "delete" + customer tid
func SignatureDeleteRequestVerify(w Workloader, pk string, sig SigningSignature) error {
	key, err := crypto.KeyFromHex(pk)
	if err != nil {
		return errors.Wrap(err, "invalid verification key")
	}

	b, err := w.SignatureChallenge()
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	if _, err := buf.WriteString("delete"); err != nil {
		return err
	}
	if _, err := buf.WriteString(fmt.Sprintf("%d", sig.Tid)); err != nil {
		return err
	}

	msg := sha256.Sum256(buf.Bytes())
	signature, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return err
	}

	return crypto.Verify(key, msg[:], signature)
}

// Verify signature
// pk is the public key used as verification key in hex encoded format
// the signature is the signature to verify (in raw binary format)
func Verify(w Workloader, pk string, sig []byte) error {
	key, err := crypto.KeyFromHex(pk)
	if err != nil {
		return errors.Wrap(err, "invalid verification key")
	}

	b, err := w.SignatureChallenge()
	if err != nil {
		return err
	}

	msg := sha256.Sum256(b)

	return crypto.Verify(key, msg[:], sig)
}
