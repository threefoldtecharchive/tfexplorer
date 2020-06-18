package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/stellar/go/xdr"
	"github.com/threefoldtech/tfexplorer/provision"
	"github.com/threefoldtech/tfexplorer/provision/builders"
	"github.com/urfave/cli"
)

func cmdsGetPool(c *cli.Context) error {
	pool, err := bcdb.Workloads.PoolGet(c.String("poolID"))
	if err != nil {
		return err
	}
	fmt.Printf("%+v \n", pool)
	return nil
}

func cmdsCreatePool(c *cli.Context) error {
	var (
		assets  = c.StringSlice("asset")
		dryRun  = c.Bool("dry-run")
		sus     = c.Uint64("sus")
		cus     = c.Uint64("cus")
		nodeIDs = c.StringSlice("nodeIDs")
		err     error
	)

	capacityBuilder := builders.NewCapacityReservationBuilder()

	capacityBuilder.
		WithSUs(sus).
		WithCUs(cus).
		WithNodeIDs(nodeIDs).
		WithCurrencies(assets)

	reservationClient := provision.NewReservationClient(bcdb, mainui)
	if dryRun {
		res, err := reservationClient.DryRunCapacity(capacityBuilder.Build(), assets)
		if err != nil {
			return errors.Wrap(err, "failed to parse reservation as JSON")
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(res)
	}

	response, err := reservationClient.DeployCapacityPool(capacityBuilder.Build(), assets)
	if err != nil {
		return errors.Wrap(err, "failed to deploy reservation")
	}

	totalAmount := xdr.Int64(0)
	for _, detail := range response.EscrowInformation.Details {
		totalAmount += detail.TotalAmount
	}

	fmt.Printf("Reservation sent to node bcdb\n")
	fmt.Printf("Resource: /reservations/pools/%v\n", response.ID)
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
