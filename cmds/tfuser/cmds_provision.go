package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/stellar/go/xdr"
	"github.com/threefoldtech/tfexplorer/provision"
	"github.com/threefoldtech/tfexplorer/provision/builders"
	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/urfave/cli"
)

var (
	day             = time.Hour * 24
	defaultDuration = day * 30
)

func cmdsProvision(c *cli.Context) error {
	var (
		d          = c.String("duration")
		assets     = c.StringSlice("asset")
		volumes    = c.StringSlice("volume")
		containers = c.StringSlice("container")
		zdbs       = c.StringSlice("zdb")
		kubes      = c.StringSlice("kube")
		networks   = c.StringSlice("network")
		qemus      = c.StringSlice("qemu")
		dryRun     = c.Bool("dry-run")
		err        error
	)

	reservationBuilder := builders.NewReservationBuilder()
	var workloadID int64 = 0

	for _, vol := range volumes {
		f, err := os.Open(vol)
		if err != nil {
			return errors.Wrap(err, "failed to open volume")
		}

		volumeBuilder, err := builders.LoadVolumeBuilder(f)
		if err != nil {
			return errors.Wrap(err, "failed to load the reservation builder")
		}
		volumeBuilder.WorkloadId = workloadID
		workloadID = +1
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
		containerBuilder.WorkloadId = workloadID
		workloadID = +1
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
		zdbBuilder.WorkloadId = workloadID
		workloadID = +1
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
		k8sBuilder.WorkloadId = workloadID
		workloadID = +1
		reservationBuilder.AddK8s(*k8sBuilder)
	}

	for _, qemu := range qemus {
		f, err := os.Open(qemu)
		if err != nil {
			return errors.Wrap(err, "failed to open qemu")
		}

		qemuBuilder, err := builders.LoadQemuBuilder(f)
		if err != nil {
			return errors.Wrap(err, "failed to load the qemu builder")
		}
		qemuBuilder.WorkloadId = workloadID
		workloadID = +1
		reservationBuilder.AddQemu(*qemuBuilder)
	}

	for _, network := range networks {
		f, err := os.Open(network)
		if err != nil {
			return errors.Wrap(err, "failed to open reservation")
		}

		networkBuilder, err := builders.LoadNetworkBuilder(f, bcdb)
		if err != nil {
			return errors.Wrap(err, "failed to load the network builder")
		}
		networkBuilder.WorkloadId = workloadID
		workloadID = +1
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

	timein := time.Now().Local().Add(duration)

	reservationBuilder.
		WithDuration(schema.Date{Time: timein}).
		WithExpirationProvisioning(schema.Date{Time: timein}).
		WithSigningRequestDeleteQuorumMin(1).
		WithSigningRequestDeleteSigners([]int64{int64(mainui.ThreebotID)})

	reservationClient := provision.NewReservationClient(bcdb, mainui)
	if dryRun {
		res, err := reservationClient.DryRun(reservationBuilder.Build(), assets)
		if err != nil {
			return errors.Wrap(err, "failed to parse reservation as JSON")
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(res)
	}

	response, err := reservationClient.Deploy(reservationBuilder.Build(), assets)
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
	reservationClient := provision.NewReservationClient(bcdb, mainui)
	return reservationClient.DeleteReservation(c.Int64("reservation"))
}
