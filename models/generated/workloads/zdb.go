package workloads

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
)

var _ Workloader = (*ZDB)(nil)
var _ Capaciter = (*ZDB)(nil)

type ZDB struct {
	ReservationInfo `bson:",inline"`

	Size            int64             `bson:"size" json:"size"`
	Mode            ZDBModeEnum       `bson:"mode" json:"mode"`
	Password        string            `bson:"password" json:"password"`
	DiskType        DiskTypeEnum      `bson:"disk_type" json:"disk_type"`
	Public          bool              `bson:"public" json:"public"`
	StatsAggregator []StatsAggregator `bson:"stats_aggregator" json:"stats_aggregator"`
}

func (z *ZDB) GetRSU() RSU {
	switch z.DiskType {
	case DiskTypeHDD:
		return RSU{
			HRU: z.Size,
		}
	case DiskTypeSSD:
		return RSU{
			SRU: z.Size,
		}
	}
	return RSU{}
}

func (v *ZDB) VerifyJSON() error {
	dup := ZDB{}

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
		return errors.New("json data does not match actual data")
	}

	return nil
}

type DiskTypeEnum uint8

const (
	DiskTypeHDD DiskTypeEnum = iota
	DiskTypeSSD
)

func (e DiskTypeEnum) String() string {
	switch e {
	case DiskTypeHDD:
		return "hdd"
	case DiskTypeSSD:
		return "ssd"
	}
	return "UNKNOWN"
}

type ZDBModeEnum uint8

const (
	ZDBModeSeq ZDBModeEnum = iota
	ZDBModeUser
)

func (e ZDBModeEnum) String() string {
	switch e {
	case ZDBModeSeq:
		return "seq"
	case ZDBModeUser:
		return "user"
	}
	return "UNKNOWN"
}
