package config

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/pkg/gridnetworks"
	"github.com/threefoldtech/tfexplorer/pkg/stellar"
)

// Settings struct
type Settings struct {
	WalletNetwork string
	TFNetwork     string
	HorizonURL    string
}

var (
	// Config is global explorer config
	Config Settings

	possibleWalletNetworks = []string{stellar.NetworkProduction}
)

// Valid checks if Config is filled with valid data
func Valid() error {
	in := func(s string, l []string) bool {
		for _, a := range l {
			if strings.EqualFold(s, a) {
				return true
			}
		}
		return false
	}
	if Config.WalletNetwork != "" && !in(Config.WalletNetwork, possibleWalletNetworks) {
		return fmt.Errorf("invalid network '%s'", Config.WalletNetwork)
	}

	if err := gridnetworks.GridNetwork(Config.TFNetwork).Valid(); err != nil {
		return errors.Wrapf(err, "invalid network")
	}

	return nil
}
