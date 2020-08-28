package workloads

import (
	"bytes"
	"fmt"
	"strings"

	schema "github.com/threefoldtech/tfexplorer/schema"
)

// ITContract contains all the detail about the IT contract of a workload
// its composed of 2 sections:
// state: is mutable and evolves during the lifetime of the workload
// contract: is the immutable part that is set at creation and never changes anymore after that
type ITContract struct {
	State    State    `bson:",inline" json:"state"`
	Contract Contract `bson:",inline" json:"contract"`
}

// GetContract implements the Workloader interface
func (c *ITContract) GetContract() *Contract { return &c.Contract }

// GetState implements the Workloader interface
func (c *ITContract) GetState() *State { return &c.State }

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

// UniqueWorkloadID was used to identify a specific workload in the legacy reservation model
// it's now there only for backwards compatibility
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
