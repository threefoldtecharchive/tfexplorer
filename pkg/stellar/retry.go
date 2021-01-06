package stellar

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/stellar/go/clients/horizonclient"
)

type (
	retryWallet struct {
		Wallet
	}
)

func (r *retryWallet) backoff(op backoff.Operation) error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = time.Minute // retry for 1 min maximum
	bo.MaxInterval = time.Second * 2

	return backoff.Retry(op, bo)
}

func (r *retryWallet) error(op string, err error) error {
	if err == nil {
		return nil
	}

	var hError horizonclient.Error
	if !errors.As(err, &hError) {
		log.Error().Err(err).Str("reason", "unknown-error-typ").Str("operation", op).Msg("operation failed permanently")
		return backoff.Permanent(err)
	}

	log.Error().
		Err(err).
		Str("operation", op).
		Str("problem", fmt.Sprintf("%+v", hError.Problem.Extras)).
		Str("status", hError.Response.Status).
		Int("status-code", hError.Problem.Status).
		Msg("operation failed")

	if hError.Response.StatusCode == http.StatusBadRequest ||
		hError.Response.StatusCode == http.StatusGatewayTimeout {
		// this error is 400 bad request is probably a problem
		// with transaction sequence number. so it's okay we retry
		return err
	}

	// otherwise
	return backoff.Permanent(err)
}

// NOTE: we don't retry the CreateAccount because it already has
// custom retry logic.
// func (r *retryWallet) CreateAccount() (encSeed string, address string, err error) {
// 	err = r.backoff(func() error {
// 		encSeed, address, err = r.Wallet.CreateAccount()
// 		return r.error("CreateAccount", err)
// 	})

// 	return
// }

func (r *retryWallet) Refund(encryptedSeed string, memo string, asset Asset) (err error) {
	err = r.backoff(func() error {
		err = r.Wallet.Refund(encryptedSeed, memo, asset)
		return r.error("Refund", err)
	})

	return
}

func (r *retryWallet) PayoutFarmers(encryptedSeed string, destinations []PayoutInfo, memo string, asset Asset) (err error) {
	err = r.backoff(func() error {
		err = r.Wallet.PayoutFarmers(encryptedSeed, destinations, memo, asset)
		return r.error("PayoutFarmers", err)
	})

	return
}
