package workloads

import (
	"encoding/json"
	"reflect"

	"github.com/rs/zerolog/log"

	"github.com/pkg/errors"
)

var _ Workloader = (*Volume)(nil)
var _ Capaciter = (*Volume)(nil)

type Volume struct {
	ReservationInfo

	Size int64          `bson:"size" json:"size"`
	Type VolumeTypeEnum `bson:"type" json:"type"`
}

func (v *Volume) GetRSU() RSU {
	switch v.Type {
	case VolumeTypeHDD:
		return RSU{
			HRU: v.Size,
		}
	case VolumeTypeSSD:
		return RSU{
			SRU: v.Size,
		}
	}
	return RSU{}
}

func (v *Volume) VerifyJSON() error {
	dup := Volume{}

	if err := json.Unmarshal([]byte(v.Json), &dup); err != nil {
		return errors.Wrap(err, "invalid json data")
	}

	// override the fields which are not part of the signature
	dup.ID = v.ID
	dup.Json = v.Json
	dup.CustomerTid = v.CustomerTid
	dup.NextAction = v.NextAction
	dup.SignaturesProvision = v.SignaturesProvision
	dup.SignatureFarmer = v.SignatureFarmer
	dup.SignaturesDelete = v.SignaturesDelete
	dup.Epoch = v.Epoch
	dup.Metadata = v.Metadata
	dup.Result = v.Result
	dup.WorkloadType = v.WorkloadType

	if match := reflect.DeepEqual(v, dup); !match {
		log.Debug().Msgf("%v", v)
		log.Debug().Msgf("%v", dup)
		return errors.New("json data does not match actual data")
	}

	return nil
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
