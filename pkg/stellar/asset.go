package stellar

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Asset on the stellar network, both code and issuer in the form <CODE>:<ISSUER>
type Asset string

// Supported assets for the wallet. Assets are different based on testnet/mainnet
const (
	TFTMainnet Asset = "TFT:GBOVQKJYHXRR3DX6NOX2RRYFRCUMSADGDESTDNBDS6CDVLGVESRTAC47"
)

// internal vars to set up the wallet with supported assets
var (
	mainnetAssets = map[Asset]struct{}{
		TFTMainnet: {},
	}
)

// Code of the asset
func (a Asset) Code() string {
	return strings.Split(string(a), ":")[0]
}

// Issuer of the asset
func (a Asset) Issuer() string {
	return strings.Split(string(a), ":")[1]
}

// String implements Stringer interface
func (a Asset) String() string {
	return string(a)
}

// Validate if the asset is entered in the expected format
func (a Asset) validate() error {
	parts := strings.Split(string(a), ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid amount of parts in asset string, got %d, expected 2", len(parts))
	}
	if len(parts[0]) == 0 {
		return errors.New("missing code in asset")
	}
	if len(parts[1]) == 0 {
		return errors.New("missing issuer in asset")
	}
	return nil
}
