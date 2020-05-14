package builders

import (
	"encoding/json"
	"io"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
)

// DebugBuilder is a struct that can set/request system debug information
type DebugBuilder struct {
	workloads.Debug
}

// NewDebugBuilder creates a new debug builder
func NewDebugBuilder(nodeID string) *DebugBuilder {
	return &DebugBuilder{
		Debug: workloads.Debug{
			NodeId:  nodeID,
			Sysdiag: false,
		},
	}
}

// LoadDebugBuilder loads a debug builder based on a file path
func LoadDebugBuilder(reader io.Reader) (*DebugBuilder, error) {
	debug := workloads.Debug{}

	err := json.NewDecoder(reader).Decode(&debug)
	if err != nil {
		return &DebugBuilder{}, err
	}

	return &DebugBuilder{Debug: debug}, nil
}

// Save saves the debug builder to an IO.Writer
func (d *DebugBuilder) Save(writer io.Writer) error {
	err := json.NewEncoder(writer).Encode(d.Debug)
	if err != nil {
		return err
	}
	return err
}

// Build does nothing for now
func (d *DebugBuilder) Build() (workloads.Debug, error) {
	return d.Debug, nil
}

// WithSysdiag enable/disable the system diagnostic request flag
func (d *DebugBuilder) WithSysdiag(enabled bool) *DebugBuilder {
	d.Debug.Sysdiag = enabled
	return d
}
