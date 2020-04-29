package main

import (
	"fmt"
	"strings"

	"github.com/threefoldtech/tfexplorer/builders"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"

	"github.com/urfave/cli"
)

func generateVolume(c *cli.Context) error {
	s := c.Int64("size")
	t := strings.ToLower(c.String("type"))

	if t != workloads.DiskTypeHDD.String() && t != workloads.DiskTypeSSD.String() {
		return fmt.Errorf("volume type can only hdd or ssd")
	}

	if s < 1 { //TODO: upper bound ?
		return fmt.Errorf("size cannot be less then 1")
	}

	var volumeType workloads.VolumeTypeEnum
	if t == workloads.DiskTypeHDD.String() {
		volumeType = workloads.VolumeTypeEnum(workloads.VolumeTypeSSD)
	} else if t == workloads.DiskTypeSSD.String() {
		volumeType = workloads.VolumeTypeEnum(workloads.VolumeTypeHDD)
	}

	volumeBuilder := builders.NewVolumeBuilder(c.String("node"), volumeType)
	volumeBuilder.WithSize(s)
	return writeWorkload(c.GlobalString("output"), volumeBuilder.Build())
}
