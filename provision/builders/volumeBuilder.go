package builders

import (
	"encoding/json"
	"io"
	"time"

	"github.com/threefoldtech/tfexplorer/models/workloads"
	"github.com/threefoldtech/tfexplorer/schema"
)

// VolumeBuilder is a struct that can build volumes
type VolumeBuilder struct {
	workloads.Volume
}

// NewVolumeBuilder creates a new volume builder
func NewVolumeBuilder(nodeID string, size int64, volumeType workloads.VolumeTypeEnum) *VolumeBuilder {
	return &VolumeBuilder{
		Volume: workloads.Volume{
			ReservationInfo: workloads.ReservationInfo{
				WorkloadId:   1,
				NodeId:       nodeID,
				WorkloadType: workloads.WorkloadTypeVolume,
			},
			Size: size,
			Type: volumeType,
		},
	}
}

// LoadVolumeBuilder loads a volume builder based on a file path
func LoadVolumeBuilder(reader io.Reader) (*VolumeBuilder, error) {
	volume := workloads.Volume{}

	err := json.NewDecoder(reader).Decode(&volume)
	if err != nil {
		return &VolumeBuilder{}, err
	}

	return &VolumeBuilder{Volume: volume}, nil
}

// Save saves the volume builder to an IO.Writer
func (v *VolumeBuilder) Save(writer io.Writer) error {
	err := json.NewEncoder(writer).Encode(v.Volume)
	if err != nil {
		return err
	}
	return err
}

// Build returns the volume
func (v *VolumeBuilder) Build() workloads.Volume {
	v.Epoch = schema.Date{Time: time.Now()}
	return v.Volume
}

// WithNodeID sets the node ID to the volume
func (v *VolumeBuilder) WithNodeID(nodeID string) *VolumeBuilder {
	v.Volume.NodeId = nodeID
	return v
}

// WithSize sets the volume size
func (v *VolumeBuilder) WithSize(size int64) *VolumeBuilder {
	v.Volume.Size = size
	return v
}

// WithType sets the volume type
func (v *VolumeBuilder) WithType(diskType workloads.VolumeTypeEnum) *VolumeBuilder {
	v.Volume.Type = diskType
	return v
}

// WithPoolID sets the poolID to the volume
func (v *VolumeBuilder) WithPoolID(poolID int64) *VolumeBuilder {
	v.Volume.PoolId = poolID
	return v
}
