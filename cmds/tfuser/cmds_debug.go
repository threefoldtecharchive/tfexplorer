package main

import (
	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/provision/builders"
	"github.com/urfave/cli"
)

func generateDebug(c *cli.Context) error {
	debugBuilder := builders.NewDebugBuilder(c.String("node"))

	debugBuilder.
		WithSysdiag(true)

	dbg, err := debugBuilder.Build()
	if err != nil {
		return errors.Wrap(err, "failed to build debug")
	}

	return writeWorkload(c.GlobalString("output"), dbg)
}
