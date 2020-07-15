package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/threefoldtech/tfexplorer/provision/builders"
	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/threefoldtech/zos/pkg/network/types"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
}

func cmdGraphNetwork(c *cli.Context) error {
	var (
		networkSchema = c.GlobalString("schema")
		err           error
	)

	if networkSchema == "" {
		return fmt.Errorf("schema name cannot be empty")
	}
	f, err := os.Open(networkSchema)
	if err != nil {
		return err
	}
	defer f.Close()

	network, err := builders.LoadNetworkBuilder(f, bcdb)
	if err != nil {
		return err
	}

	outfile, err := os.Create(networkSchema + ".dot")
	if err != nil {
		return err
	}

	return network.NetworkGraph(outfile)
}

func cmdCreateNetwork(c *cli.Context) error {
	name := c.String("name")
	if name == "" {
		return fmt.Errorf("network name cannot be empty")
	}
	ipRange := c.String("cidr")
	if ipRange == "" {
		return fmt.Errorf("ip range cannot be empty")
	}

	ipnet, err := types.ParseIPNet(ipRange)
	if err != nil {
		errors.Wrap(err, "invalid ip range")
	}

	networkBuilder := builders.NewNetworkBuilder(name, schema.IPRange{IPNet: ipnet.IPNet}, bcdb)

	return writeWorkload(c.GlobalString("schema"), networkBuilder)
}

func cmdsAddNode(c *cli.Context) error {
	var (
		networkSchema = c.GlobalString("schema")

		nodeID = c.String("node")
		subnet = c.String("subnet")
		port   = c.Uint("port")

		forceHidden = c.Bool("force-hidden")
	)

	if networkSchema == "" {
		return fmt.Errorf("schema name cannot be empty")
	}
	f, err := os.OpenFile(c.GlobalString("schema"), os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	defer f.Close()

	network, err := builders.LoadNetworkBuilder(f, bcdb)
	if err != nil {
		return err
	}

	_, err = network.AddNode(nodeID, subnet, port, forceHidden)
	if err != nil {
		return errors.Wrapf(err, "failed to add a node to the network %s", network.Name)
	}

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return network.Save(f)
}

func cmdsAddAccess(c *cli.Context) error {
	var (
		networkSchema = c.GlobalString("schema")

		nodeID   = c.String("node")
		subnet   = c.String("subnet")
		wgPubKey = c.String("wgpubkey")

		ip4 = c.Bool("ip4")
	)

	if networkSchema == "" {
		return fmt.Errorf("schema name cannot be empty")
	}

	f, err := os.OpenFile(c.GlobalString("schema"), os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	defer f.Close()

	network, err := builders.LoadNetworkBuilder(f, bcdb)
	if err != nil {
		return err
	}

	if nodeID == "" {
		return fmt.Errorf("nodeID cannot be empty")
	}
	if subnet == "" {
		return fmt.Errorf("subnet cannot be empty")
	}

	ipnet, err := types.ParseIPNet(subnet)
	if err != nil {
		return errors.Wrap(err, "invalid subnet")
	}

	_, wgSchema, err := network.AddAccess(nodeID, schema.IPRange{IPNet: ipnet.IPNet}, wgPubKey, ip4)
	if err != nil {
		return err
	}

	fmt.Println(wgSchema)

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return network.Save(f)
}

func cmdsRemoveNode(c *cli.Context) error {
	var (
		networkSchema = c.GlobalString("schema")
		nodeID        = c.String("node")
	)

	if networkSchema == "" {
		return fmt.Errorf("schema name cannot be empty")
	}
	f, err := os.OpenFile(c.GlobalString("schema"), os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	defer f.Close()

	network, err := builders.LoadNetworkBuilder(f, bcdb)
	if err != nil {
		return err
	}

	if err := network.RemoveNode(networkSchema, nodeID); err != nil {
		return errors.Wrapf(err, "failed to remove node %s", nodeID)
	}

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return network.Save(f)
}
