package main

import (
	"encoding/binary"
	"net"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/tfexplorer/provision/builders"
	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/urfave/cli"
)

func generateVM(c *cli.Context) error {
	var (
		flist          = c.String("flist")
		hubUrl         = c.String("hub-url")
		netID          = c.String("network-id")
		ipString       = c.String("ip")
		nodeID         = c.String("node")
		publicIPString = c.String("public-ip")
		sshKeys        = c.StringSlice("ssh-keys")
		cap            = workloads.VMCapacity{
			Cpu:      c.Int64("cpu"),
			Memory:   c.Int64("memory"),
			DiskSize: c.Uint64("disk-size"),
		}
		// why is there a hub url
	)

	if netID == "" {
		return errors.New("vm requires a network to run in")
	}

	ip := net.ParseIP(ipString)
	if ip.To4() == nil {
		return errors.New("bad IP for vm")
	}
	pubip := net.ParseIP(publicIPString)
	if pubip.To4() == nil {
		return errors.New("bad public ip for vm, use x.x.x.x")
	}
	publicIP := schema.ID(ip2int(pubip))

	vm := builders.NewVMBuilder(nodeID, netID, ip)

	vm.
		WithFlist(flist).
		WithHubURL(hubUrl).
		WithPublicIP(publicIP).
		WithSSHKeys(sshKeys).
		WithVMCapacity(cap)

	if c.Int64("poolID") != 0 {
		vm.WithPoolID(c.Int64("poolID"))
	}

	return writeWorkload(c.GlobalString("schema"), vm.Build())
}

func ip2int(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}
