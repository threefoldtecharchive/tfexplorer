package workloads

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
