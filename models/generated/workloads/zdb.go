package workloads

import schema "github.com/threefoldtech/tfexplorer/schema"

type ZDB struct {
	WorkloadId      int64             `bson:"workload_id" json:"workload_id"`
	NodeId          string            `bson:"node_id" json:"node_id"`
	Size            int64             `bson:"size" json:"size"`
	Mode            ZDBModeEnum       `bson:"mode" json:"mode"`
	Password        string            `bson:"password" json:"password"`
	DiskType        DiskTypeEnum      `bson:"disk_type" json:"disk_type"`
	Public          bool              `bson:"public" json:"public"`
	StatsAggregator []StatsAggregator `bson:"stats_aggregator" json:"stats_aggregator"`
	FarmerTid       int64             `bson:"farmer_tid" json:"farmer_tid"`
	PoolId          int64             `bson:"pool_id" json:"pool_id"`

	ID                  schema.ID          `bson:"_id" json:"id"`
	Json                string             `bson:"json" json:"json"`
	CustomerTid         int64              `bson:"customer_tid" json:"customer_tid"`
	CustomerSignature   string             `bson:"customer_signature" json:"customer_signature"`
	NextAction          NextActionEnum     `bson:"next_action" json:"next_action"`
	SignaturesProvision []SigningSignature `bson:"signatures_provision" json:"signatures_provision"`
	SignaturesFarmer    []SigningSignature `bson:"signatures_farmer" json:"signatures_farmer"`
	SignaturesDelete    []SigningSignature `bson:"signatures_delete" json:"signatures_delete"`
	Epoch               schema.Date        `bson:"epoch" json:"epoch"`
	Metadata            string             `bson:"metadata" json:"metadata"`
	Result              Result             `bson:"result" json:"result"`
	WorkloadType        WorkloadTypeEnum   `bson:"workload_type" json:"workload_type"`
}

func (z ZDB) WorkloadID() int64 {
	return z.WorkloadId
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
