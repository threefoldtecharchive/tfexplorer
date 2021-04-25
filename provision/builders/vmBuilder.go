package builders

import (
	"encoding/json"
	"io"
	"net"
	"time"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/tfexplorer/schema"
)

// VMBuilder is a struct that can build K8S's
type VMBuilder struct {
	workloads.VirtualMachine
}

// NewVMBuilder creates a new VM builder
func NewVMBuilder(nodeID, networkID string, IP net.IP) *VMBuilder {
	return &VMBuilder{
		VirtualMachine: workloads.VirtualMachine{
			ReservationInfo: workloads.ReservationInfo{
				WorkloadId:   1,
				NodeId:       nodeID,
				WorkloadType: workloads.WorkloadTypeKubernetes,
			},
			NetworkId: networkID,
			Ipaddress: IP,
		},
	}
}

// LoadVMBuilder loads a vm builder based on a file path
func LoadVMBuilder(reader io.Reader) (*VMBuilder, error) {
	vm := workloads.VirtualMachine{}

	err := json.NewDecoder(reader).Decode(&vm)
	if err != nil {
		return &VMBuilder{}, err
	}

	return &VMBuilder{VirtualMachine: vm}, nil
}

// Save saves the VM builder to an IO.Writer
func (vm *VMBuilder) Save(writer io.Writer) error {
	err := json.NewEncoder(writer).Encode(vm.VirtualMachine)
	if err != nil {
		return err
	}
	return err
}

// Build returns the kubernetes
func (vm *VMBuilder) Build() workloads.VirtualMachine {
	vm.Epoch = schema.Date{Time: time.Now()}
	return vm.VirtualMachine
}

// WithNodeID sets the node ID to the K8S
func (vm *VMBuilder) WithNodeID(nodeID string) *VMBuilder {
	vm.NodeId = nodeID
	return vm
}

// WithNetworkID sets the network id to the K8S
func (vm *VMBuilder) WithNetworkID(id string) *VMBuilder {
	vm.VirtualMachine.NetworkId = id
	return vm
}

// WithIPAddress sets the ip address to the K8S
func (vm *VMBuilder) WithIPAddress(ip net.IP) *VMBuilder {
	vm.VirtualMachine.Ipaddress = ip
	return vm
}

// WithSSHKeys sets the ssh keys to the K8S
func (vm *VMBuilder) WithSSHKeys(sshKeys []string) *VMBuilder {
	vm.VirtualMachine.SshKeys = sshKeys
	return vm
}

// WithPoolID sets the poolID to the k8s
func (vm *VMBuilder) WithPoolID(poolID int64) *VMBuilder {
	vm.PoolId = poolID
	return vm
}

// WithContainerCapacity sets the container capacity to the container
func (vm *VMBuilder) WithVMCapacity(cap workloads.VMCapacity) *VMBuilder {
	vm.VirtualMachine.Capacity = cap
	return vm
}

// WithHubURL sets the hub url to the vm
func (vm *VMBuilder) WithHubURL(url string) *VMBuilder {
	vm.VirtualMachine.HubUrl = url
	return vm
}

// WithFlist sets the flist to the vm
func (vm *VMBuilder) WithFlist(flist string) *VMBuilder {
	vm.VirtualMachine.Flist = flist
	return vm
}

// WithFlist sets the flist to the vm
func (vm *VMBuilder) WithPublicIP(publicIP schema.ID) *VMBuilder {
	vm.VirtualMachine.PublicIP = publicIP
	return vm
}
