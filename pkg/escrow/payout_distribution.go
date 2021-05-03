package escrow

import (
	"fmt"
)

const (

	// WisdomWallet as defined in https://wiki.threefold.io/#/threefold_foundation_wallets
	WisdomWallet = "GAI4C2BGOA3YHVQZZW7OW4FHOGGYWTUBEVNHB6MW4ZAFG7ZAA7D5IPC3"
)

type PaymentDestination uint8

const (
	FarmerDestination     PaymentDestination = 0
	BurnedDestination     PaymentDestination = 1
	FoundationDestination PaymentDestination = 2
	SalesDestination      PaymentDestination = 3
	WisdomDestination     PaymentDestination = 4
)

type PaymentDistribution map[PaymentDestination]uint8

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
	DistributionV2             = "grid2"
	DistributionV3             = "grid3"
	DistributionCertifiedSales = "certified-sales"
	DistributionFamerSales     = "farmer-sales"
)

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

type Payout struct {
	Destination  PaymentDestination
	Distribution uint8
	Address      string
}

func (p *Payout) Valid() error {
	if p.Distribution == 0 {
		return nil
	}

	if len(p.Address) == 0 {
		return fmt.Errorf("missing address for payout")
	}

	return nil
}
