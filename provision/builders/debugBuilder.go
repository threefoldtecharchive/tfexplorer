package builders

import (
	"encoding/json"
	_ "fmt"
	"io"

	_ "github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
)

// ContainerBuilder is a struct that can build containers
type DebugBuilder struct {
	workloads.Debug
}

// NewContainerBuilder creates a new container builder
func NewDebugBuilder(nodeID string) *DebugBuilder {
	return &DebugBuilder{
		Debug: workloads.Debug{
			NodeId:  nodeID,
			Sysdiag: false,
		},
	}
}

// LoadContainerBuilder loads a container builder based on a file path
func LoadDebugBuilder(reader io.Reader) (*DebugBuilder, error) {
	debug := workloads.Debug{}

	err := json.NewDecoder(reader).Decode(&debug)
	if err != nil {
		return &DebugBuilder{}, err
	}

	return &DebugBuilder{Debug: debug}, nil
}

// Save saves the container builder to an IO.Writer
func (d *DebugBuilder) Save(writer io.Writer) error {
	err := json.NewEncoder(writer).Encode(d.Debug)
	if err != nil {
		return err
	}
	return err
}

// Build validates and encrypts the secret environment of the container
func (d *DebugBuilder) Build() (workloads.Debug, error) {
	return d.Debug, nil
}

// WithSysdiag sets the system diagnostic request flag
func (d *DebugBuilder) WithSysdiag(enabled bool) *DebugBuilder {
	d.Debug.Sysdiag = enabled
	return d
}
