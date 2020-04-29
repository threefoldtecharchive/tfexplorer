package main

import (
	"net"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/builders"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/urfave/cli"
)

func generateContainer(c *cli.Context) error {
	cap := workloads.ContainerCapacity{
		Cpu:    c.Int64("cpu"),
		Memory: c.Int64("memory"),
	}

	network := []workloads.NetworkConnection{
		workloads.NetworkConnection{
			NetworkId: c.String("network"),
			Ipaddress: net.ParseIP(c.String("ip")),
			PublicIp6: c.Bool("public6"),
		},
	}

	containerBuilder := builders.NewContainerBuilder(c.String("node"), c.String("flist"), cap, network)

	containerBuilder, err := containerBuilder.WithEnvs(c.StringSlice("envs"))
	if err != nil {
		return err
	}
	containerBuilder, err = containerBuilder.WithVolumes(c.StringSlice("mounts"))
	if err != nil {
		return err
	}
	containerBuilder, err = containerBuilder.WithLogs(c.String("stdout"), c.String("stderr"))
	if err != nil {
		return err
	}
	containerBuilder, err = containerBuilder.WithStatsAggregator(c.String("stats"))
	if err != nil {
		return err
	}

	containerBuilder.WithEntrypoint(c.String("entrypoint")).WithInteractive(c.Bool("corex"))

	container, err := containerBuilder.Build()
	if err != nil {
		return errors.Wrap(err, "failed to build container")
	}
	return writeWorkload(c.GlobalString("output"), container)
}
