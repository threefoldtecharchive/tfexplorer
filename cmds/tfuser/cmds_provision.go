package main

import (
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/stellar/go/xdr"
	"github.com/threefoldtech/tfexplorer/builders"

	"github.com/urfave/cli"
)

var (
	day             = time.Hour * 24
	defaultDuration = day * 30
)

func cmdsProvision(c *cli.Context) error {
	var (
		seedPath   = mainSeed
		d          = c.String("duration")
		assets     = c.StringSlice("asset")
		volumes    = c.StringSlice("volume")
		containers = c.StringSlice("container")
		zdbs       = c.StringSlice("zdb")
		kubes      = c.StringSlice("kube")
		networks   = c.StringSlice("network")
		dryRun     = c.Bool("dry-run")
		err        error
	)

	reservationBuilder := builders.NewReservationBuilder(bcdb, mainui)

	for _, vol := range volumes {
		f, err := os.Open(vol)
		if err != nil {
			return errors.Wrap(err, "failed to open volume")
		}

		volumeBuilder, err := builders.LoadVolumeBuilder(f)
		if err != nil {
			return errors.Wrap(err, "failed to load the reservation builder")
		}
		reservationBuilder.AddVolume(*volumeBuilder)
	}

	for _, cont := range containers {
		f, err := os.Open(cont)
		if err != nil {
			return errors.Wrap(err, "failed to open container")
		}

		containerBuilder, err := builders.LoadContainerBuilder(f)
		if err != nil {
			return errors.Wrap(err, "failed to load the reservation builder")
		}
		reservationBuilder.AddContainer(*containerBuilder)
	}

	for _, zdb := range zdbs {
		f, err := os.Open(zdb)
		if err != nil {
			return errors.Wrap(err, "failed to open zdb")
		}

		zdbBuilder, err := builders.LoadZdbBuilder(f)
		if err != nil {
			return errors.Wrap(err, "failed to load the zdb builder")
		}
		reservationBuilder.AddZdb(*zdbBuilder)
	}

	for _, k8s := range kubes {
		f, err := os.Open(k8s)
		if err != nil {
			return errors.Wrap(err, "failed to open kube")
		}

		k8sBuilder, err := builders.LoadK8sBuilder(f)
		if err != nil {
			return errors.Wrap(err, "failed to load the k8s builder")
		}
		reservationBuilder.AddK8s(*k8sBuilder)
	}

	for _, network := range networks {
		f, err := os.Open(network)
		if err != nil {
			return errors.Wrap(err, "failed to open reservation")
		}

		networkBuilder, err := builders.LoadNetworkBuilder(f)
		if err != nil {
			return errors.Wrap(err, "failed to load the network builder")
		}
		reservationBuilder.AddNetwork(*networkBuilder)
	}

	var duration time.Duration
	if d == "" {
		duration = defaultDuration
	} else {
		duration, err = time.ParseDuration(d)
		if err != nil {
			nrDays, err := strconv.Atoi(d)
			if err != nil {
				return errors.Wrap(err, "unsupported duration format")
			}
			duration = time.Duration(nrDays) * day
		}
	}

	reservationBuilder.WithDuration(duration).WithDryRun(dryRun).WithSeedPath(seedPath).WithAssets(assets)

	response, err := reservationBuilder.Deploy()
	if err != nil {
		return errors.Wrap(err, "failed to deploy reservation")
	}

	totalAmount := xdr.Int64(0)
	for _, detail := range response.EscrowInformation.Details {
		totalAmount += detail.TotalAmount
	}

	fmt.Printf("Reservation for %v send to node bcdb\n", d)
	fmt.Printf("Resource: /reservations/%v\n", response.ID)
	fmt.Println()

	fmt.Printf("Reservation id: %d \n", response.ID)
	fmt.Printf("Asset to pay: %s\n", response.EscrowInformation.Asset)
	fmt.Printf("Reservation escrow address: %s \n", response.EscrowInformation.Address)
	fmt.Printf("Reservation amount: %s %s\n", formatCurrency(totalAmount), response.EscrowInformation.Asset.Code())

	for _, detail := range response.EscrowInformation.Details {
		fmt.Println()
		fmt.Printf("FarmerID: %v\n", detail.FarmerID)
		fmt.Printf("Amount: %s\n", formatCurrency(detail.TotalAmount))
	}

	return nil
}

func formatCurrency(amount xdr.Int64) string {
	currency := big.NewRat(int64(amount), 1e7)
	return currency.FloatString(7)
}

func cmdsDeleteReservation(c *cli.Context) error {
	reservationBuilder := builders.NewReservationBuilder(bcdb, mainui)
	return reservationBuilder.DeleteReservation(c.Int64("reservation"))
}
