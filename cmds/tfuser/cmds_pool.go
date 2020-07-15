package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/pkg/capacity/types"
	"github.com/threefoldtech/tfexplorer/provision"
	"github.com/threefoldtech/tfexplorer/provision/builders"
	"github.com/urfave/cli"
)

func cmdsGetPool(c *cli.Context) error {
	pool, err := bcdb.Workloads.PoolGet(c.String("poolID"))
	if err != nil {
		return err
	}

	fmt.Println(formatPool(pool))
	return nil
}

func cmdsGetPoolsByOwner(c *cli.Context) error {
	userID := c.String("ownerID")
	if userID == "" && mainui != nil {
		userID = fmt.Sprintf("%d", mainui.ThreebotID)
	}

	pools, err := bcdb.Workloads.PoolsGetByOwner(userID)
	if err != nil {
		return err
	}
	for _, p := range pools {
		fmt.Println(formatPool(p))
		fmt.Println()
	}
	return nil
}

func cmdsCreatePool(c *cli.Context) error {
	var (
		assets  = c.StringSlice("asset")
		dryRun  = c.Bool("dry-run")
		sus     = c.Uint64("sus")
		cus     = c.Uint64("cus")
		nodeIDs = c.StringSlice("nodeIDs")
		poolID  = c.Int64("poolID")
		err     error
	)

	if len(nodeIDs) == 0 {
		return fmt.Errorf("missing --nodeIDs flag: a pool required at least one node")
	}

	if len(assets) == 0 {
		assets = append(assets, "TFT")
	}

	capacityBuilder := builders.NewCapacityReservationBuilder()

	capacityBuilder.
		WithSUs(sus).
		WithCUs(cus).
		WithNodeIDs(nodeIDs).
		WithCurrencies(assets)

	if poolID != 0 {
		capacityBuilder.WithPoolID(poolID)
	}

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

	fmt.Printf("Pool reservation sent to node bcdb\n")
	fmt.Printf("Resource: /reservations/pools/%v\n", response.ID)
	fmt.Println()

	fmt.Printf("Capacity reservation id: %d \n", response.ID)
	fmt.Printf("Asset to pay: %s\n", response.EscrowInformation.Asset)
	fmt.Printf("Reservation escrow address: %s \n", response.EscrowInformation.Address)
	fmt.Printf("Reservation amount: %s %s\n", formatCurrency(response.EscrowInformation.Amount), response.EscrowInformation.Asset.Code())

	return nil
}

func formatPool(pool types.Pool) string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "Pool ID: %d\n", pool.ID)
	fmt.Fprintf(b, "Capacity left:\n")
	fmt.Fprintf(b, "  Compute unit: %.2f\n", pool.Cus)
	fmt.Fprintf(b, "  Storage unit: %.2f\n", pool.Sus)
	fmt.Fprintf(b, "Capacity in usage:\n")
	fmt.Fprintf(b, "  Compute unit: %.2f\n", pool.ActiveCU)
	fmt.Fprintf(b, "  Storage unit: %.2f\n", pool.ActiveSU)
	fmt.Fprintf(b, "Will expired at: %s\n", time.Unix(pool.EmptyAt, 0).Format(time.RFC1123))
	return b.String()
}
