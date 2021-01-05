// Package gridnetworks is a separate package to avoid import cycles due to a wild
// global config
package gridnetworks

import "fmt"

// GridNetwork type
type GridNetwork string

var (
	// GridNetworkMainnet is the grid mainnet
	GridNetworkMainnet GridNetwork = "mainnet"
	// GridNetworkTestnet is the grid testnet
	GridNetworkTestnet GridNetwork = "testnet"
	// GridNetworkDevnet is the grid devnet
	GridNetworkDevnet GridNetwork = "devnet"
)

// Divisor gets a divisor for the fee to be paid based on the current
// grid network
func (g GridNetwork) Divisor() (int64, error) {
	switch g {
	case GridNetworkMainnet:
		return 1, nil
	case GridNetworkTestnet:
		return 10, nil
	case GridNetworkDevnet:
		return 100, nil
	default:
		return 0, fmt.Errorf("unknown grid network")
	}
}

func (g GridNetwork) Valid() error {
	_, err := g.Divisor()
	return err
}
