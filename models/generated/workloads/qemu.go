package workloads

import "net"

type Qemu struct {
	WorkloadId int64        `bson:"workload_id" json:"workload_id"`
	NodeId     string       `bson:"node_id" json:"node_id"`
	Ipaddress  net.IP       `bson:"ipaddress" json:"ipaddress"`
	Image      string       `bson:"image" json:"image"`
	Capacity   QemuCapacity `bson:"capacity" json:"capacity"`
}

func (q Qemu) WorkloadID() int64 {
	return q.WorkloadId
}

// QemuCapacity is the amount of resource to allocate to the virtual machine
type QemuCapacity struct {
	// Number of CPU
	CPU uint `json:"cpu"`
	// Memory in MiB
	Memory uint64 `json:"memory"`
	// HDD in GB
	HDDSize uint64 `json:"hdd"`
}
