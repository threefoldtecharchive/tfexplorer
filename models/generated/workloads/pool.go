package workloads

import schema "github.com/threefoldtech/tfexplorer/schema"

type CapacityPool struct {
	ID           schema.ID `bson:"_id" json:"id"`
	WorkloadId   int64     `bson:"workload_id" json:"workload_id"`
	UsedCapacity Capacity  `bson:"used_capacity" json:"used_capacity"`
	NodeIDs      []string  `bson:"node_ids" json:"node_ids"`
}

type Capacity struct {
	CPU int64 `bson:"cpu" json:"cpu"`
	MRU int64 `bson:"mru" json:"mru"`
	HRU int64 `bson:"hru" json:"hru"`
	SRU int64 `bson:"sru" json:"sru"`
}

func (c CapacityPool) WorkloadID() int64 {
	return c.WorkloadId
}
