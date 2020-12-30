package config

import (
	"fmt"
	"strings"

	"github.com/threefoldtech/tfexplorer/pkg/gridnetworks"
	"github.com/threefoldtech/tfexplorer/pkg/stellar"
)

// Settings struct
type Settings struct {
	WalletNetwork string
	TFNetwork     string
}

var (
	// Config is global explorer config
	Config Settings

	possibleWalletNetworks = []string{stellar.NetworkProduction}
	possibleGridNetworks   = []string{gridnetworks.GridNetworkMainnet,
		gridnetworks.GridNetworkTestnet, gridnetworks.GridNetworkDevnet}
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
	if !in(Config.TFNetwork, possibleGridNetworks) {
		return fmt.Errorf("invalid network '%s'", Config.TFNetwork)
	}

	return nil
}
