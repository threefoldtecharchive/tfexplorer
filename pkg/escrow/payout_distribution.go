package escrow

import (
	"fmt"
)

const (

	// WisdomWallet as defined in https://wiki.threefold.io/#/threefold_foundation_wallets
	WisdomWallet = "GAI4C2BGOA3YHVQZZW7OW4FHOGGYWTUBEVNHB6MW4ZAFG7ZAA7D5IPC3"
)

// PaymentDestination type
type PaymentDestination uint8

const (
	// FarmerDestination destination
	FarmerDestination PaymentDestination = 0
	// BurnedDestination destination
	BurnedDestination PaymentDestination = 1
	// FoundationDestination destination
	FoundationDestination PaymentDestination = 2
	// SalesDestination destination
	SalesDestination PaymentDestination = 3
	// WisdomDestination destination
	WisdomDestination PaymentDestination = 4
)

// PaymentDistribution type is map from destination to a percent
type PaymentDistribution map[PaymentDestination]uint8

// Valid checks if distribution is valid
func (p PaymentDistribution) Valid() error {
	var total uint8
	for _, v := range p {
		total += v
	}

	if total != 100 {
		return fmt.Errorf("expected total payout distribution to be 100, got %d", total)
	}

	return nil
}

const (
	// DistributionV2 uses legacy v2.x distribution
	DistributionV2 = "grid2"
	// DistributionV3 uses new v3.0 distribution
	DistributionV3 = "grid3"
	// DistributionCertifiedSales uses distribution if capacity is sold over
	// certified sales channel
	DistributionCertifiedSales = "certified-sales"
	// DistributionFamerSales uses distribution if farmer is re-buying or selling
	// his own capacity
	DistributionFamerSales = "farmer-sales"
)

// AssetDistributions map
var AssetDistributions = map[string]PaymentDistribution{
	DistributionV2: {
		FarmerDestination:     90,
		FoundationDestination: 10,
	},
	DistributionV3: {
		FarmerDestination:     10,
		BurnedDestination:     40,
		FoundationDestination: 10,
		WisdomDestination:     40,
	},
	DistributionCertifiedSales: {
		FarmerDestination:     10,
		BurnedDestination:     25,
		FoundationDestination: 10,
		SalesDestination:      55,
	},
	DistributionFamerSales: {
		FarmerDestination:     70,
		BurnedDestination:     25,
		FoundationDestination: 5,
	},
}

// Payout structure
type Payout struct {
	Destination  PaymentDestination
	Distribution uint8
	Address      string
}

// Valid checks if payout is valid,
func (p *Payout) Valid() error {
	if p.Distribution == 0 {
		return nil
	}

	if len(p.Address) == 0 {
		return fmt.Errorf("missing address for payout")
	}

	return nil
}
