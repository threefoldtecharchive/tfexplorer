package workloads

type Debug struct {
	WorkloadId int64  `bson:"workload_id" json:"workload_id"`
	NodeId     string `bson:"node_id" json:"node_id"`
	FarmerTid  int64  `bson:"farmer_tid" json:"farmer_tid"`
	Sysdiag    bool   `bson:"sysdiag" json:"sysdiag"`
}

func (d Debug) WorkloadID() int64 {
	return d.WorkloadId
}
