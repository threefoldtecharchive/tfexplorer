package main

import (
	"net"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/tfexplorer/provision/builders"
	"github.com/urfave/cli"
)

func generateQemu(c *cli.Context) error {
	var (
		nodeID            = c.String("node")
		ipString          = c.String("ip")
		image             = c.String("image")
		imageFlistStorage = c.String("image-flist-storage")
	)

	ip := net.ParseIP(ipString)
	if ip.To4() == nil {
		return errors.New("bad IP for vm")
	}

	if image == "" {
		return errors.New("vm requires a image to boot from")
	}

	if imageFlistStorage == "" {
		return errors.New("imageFlistStorage is required")
	}

	cap := workloads.QemuCapacity{
		CPU:     c.Uint("cpu"),
		Memory:  c.Uint64("memory"),
		HDDSize: c.Uint64("hddsize"),
	}

	qemu := builders.NewQemuBuilder(nodeID, ip, image, imageFlistStorage, cap)
	return writeWorkload(c.GlobalString("schema"), qemu.Build())
}