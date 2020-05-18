package builders

import (
	"encoding/json"
	"io"
	"net"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
)

// QemuBuilder is a struct that can build K8S's
type QemuBuilder struct {
	workloads.Qemu
}

// NewQemuBuilder creates a new Qemu builder
func NewQemuBuilder(nodeID string, networkID string, IP net.IP, image string, capacity workloads.QemuCapacity) *QemuBuilder {
	return &QemuBuilder{
		Qemu: workloads.Qemu{
			NodeId:    nodeID,
			NetworkId: networkID,
			Ipaddress: IP,
			Image:     image,
			Capacity:  capacity,
		},
	}
}

// LoadQemuBuilder loads a qemu builder based on a file path
func LoadQemuBuilder(reader io.Reader) (*QemuBuilder, error) {
	qemu := workloads.Qemu{}

	err := json.NewDecoder(reader).Decode(&qemu)
	if err != nil {
		return &QemuBuilder{}, err
	}

	return &QemuBuilder{Qemu: qemu}, nil
}

// Save saves the Qemu builder to an IO.Writer
func (qemu *QemuBuilder) Save(writer io.Writer) error {
	err := json.NewEncoder(writer).Encode(qemu.Qemu)
	if err != nil {
		return err
	}
	return err
}

// Build returns the qemu
func (qemu *QemuBuilder) Build() workloads.Qemu {
	return qemu.Qemu
}

// WithNodeID sets the node ID to the Qemu
func (qemu *QemuBuilder) WithNodeID(nodeID string) *QemuBuilder {
	qemu.Qemu.NodeId = nodeID
	return qemu
}

// WithIPAddress sets the ip address to the Qemu
func (qemu *QemuBuilder) WithIPAddress(ip net.IP) *QemuBuilder {
	qemu.Qemu.Ipaddress = ip
	return qemu
}

// WithNetworkID sets the ip address to the Qemu
func (qemu *QemuBuilder) WithNetworkID(netID string) *QemuBuilder {
	qemu.Qemu.NetworkId = netID
	return qemu
}

// WithImage sets the image id to the Qemu
func (qemu *QemuBuilder) WithImage(image string) *QemuBuilder {
	qemu.Qemu.Image = image
	return qemu
}

// WithCapacity sets the capacity id to the Qemu
func (qemu *QemuBuilder) WithCapacity(capacity workloads.QemuCapacity) *QemuBuilder {
	qemu.Qemu.Capacity = capacity
	return qemu
}
