package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/stellar/go/xdr"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	wrklds "github.com/threefoldtech/tfexplorer/pkg/workloads"

	"github.com/threefoldtech/tfexplorer/provision"
	"github.com/urfave/cli"
)

var (
	day             = time.Hour * 24
	defaultDuration = day * 30
)

func cmdsProvision(c *cli.Context) error {
	var (
		d           = c.String("duration")
		assets      = c.StringSlice("asset")
		workloaders = c.StringSlice("workload")
		dryRun      = c.Bool("dry-run")
		err         error
	)

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

	reservationClient := provision.NewReservationClient(bcdb, mainui)

	results := make([]wrklds.ReservationCreateResponse, 0, len(workloaders))
	for _, workload := range workloaders {
		buffer, err := ioutil.ReadFile(workload)
		if err != nil {
			return errors.Wrap(err, "failed to read workload")
		}

		workloader, err := workloads.UnmarshalJSON(buffer)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal json to workload")
		}

		if dryRun {
			res, err := reservationClient.DryRun(workloader, timein)
			if err != nil {
				return errors.Wrap(err, "failed to parse reservation as JSON")
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err = enc.Encode(res); err != nil {
				return errors.Wrap(err, "failed to encode reservation")
			}
			continue
		}

		result, err := reservationClient.Deploy(workloader, assets, timein)
		if err != nil {
			return errors.Wrap(err, "failed to deploy reservation")
		}
		results = append(results, result)
	}

	for _, r := range results {
		fmt.Printf("workloads send: ID %d\n", r.ID)
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
