package workloads

import (
	"bytes"
	"fmt"
)

var _ Workloader = (*Volume)(nil)
var _ Capaciter = (*Volume)(nil)

type Volume struct {
	contract Contract
	state    State

	Size int64          `bson:"size" json:"size"`
	Type VolumeTypeEnum `bson:"type" json:"type"`
}

// Contract implements the Workloader interface
func (v *Volume) Contract() Contract { return v.contract }

// State implements the Workloader interface
func (v *Volume) State() *State { return &v.state }

// GetRSU implements the Capaciter interface
func (v *Volume) GetRSU() RSU {
	switch v.Type {
	case VolumeTypeHDD:
		return RSU{
			HRU: float64(v.Size),
		}
	case VolumeTypeSSD:
		return RSU{
			SRU: float64(v.Size),
		}
	}
	return RSU{}
}

func (v *Volume) SignatureChallenge() ([]byte, error) {
	ric, err := v.contract.SignatureChallenge()
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer(ric)
	fmt.Fprintf(b, "%d", v.Size)
	fmt.Fprintf(b, "%s", v.Type.String())

	return b.Bytes(), nil
}

type VolumeTypeEnum uint8

const (
	VolumeTypeHDD VolumeTypeEnum = iota
	VolumeTypeSSD
)

func (e VolumeTypeEnum) String() string {
	switch e {
	case VolumeTypeHDD:
		return "HDD"
	case VolumeTypeSSD:
		return "SSD"
	}
	return "UNKNOWN"
}
