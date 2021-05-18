package escrow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPayoutDistributionValidation(t *testing.T) {
	distributions := []PaymentDistribution{
		{
			FarmerDestination:     23,
			BurnedDestination:     56,
			FoundationDestination: 0,
		},
		{
			FarmerDestination:     33,
			BurnedDestination:     33,
			FoundationDestination: 33,
		},
		{
			FarmerDestination:     50,
			BurnedDestination:     40,
			FoundationDestination: 10,
		},
	}

	assert.Error(t, distributions[0].Valid(), "")
	assert.Error(t, distributions[1].Valid(), "")
	assert.NoError(t, distributions[2].Valid())
}

func TestKnownPayoutDistributions(t *testing.T) {
	for _, pd := range AssetDistributions {
		assert.NoError(t, pd.Valid())
	}
}
