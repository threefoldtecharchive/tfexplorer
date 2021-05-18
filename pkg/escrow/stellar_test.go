package escrow

import (
	"fmt"
	"testing"

	"github.com/stellar/go/xdr"
	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/tfexplorer/pkg/gridnetworks"
	"github.com/threefoldtech/tfexplorer/pkg/stellar"
)

func TestPayoutDistribution(t *testing.T) {
	// for these tests, keep in mind that the `amount` given is in the highest
	// precision of the underlying wallet, but the reservation costs only have
	// up to 6 digits precision. In case of stellar, the wallet has 7 digits precision.
	// This means the smallest amount that will be expressed is `10` rather than `1`.
	//
	// note that the actual amount to be paid can have up to the wallets precision,
	// i.e. it is possible to have greater than 6 digits precision
	pds := []PaymentDistribution{
		{
			FarmerDestination:     50,
			BurnedDestination:     50,
			FoundationDestination: 0,
		},
		{
			FarmerDestination:     34,
			BurnedDestination:     33,
			FoundationDestination: 33,
		},
		{
			FarmerDestination:     40,
			BurnedDestination:     40,
			FoundationDestination: 20,
		},
		{
			FarmerDestination:     0,
			BurnedDestination:     73,
			FoundationDestination: 27,
		},
	}

	payouts := []struct {
		payouts []Payout
		inputs  []int64
		output  [][]int64
	}{
		{
			payouts: []Payout{
				{Destination: FarmerDestination, Distribution: 50},
				{Destination: BurnedDestination, Distribution: 50},
				{Destination: FoundationDestination, Distribution: 0},
			},
			inputs: []int64{10, 330},
			output: [][]int64{{5, 5, 0}, {165, 165, 0}},
		},
		{
			payouts: []Payout{
				{Destination: FarmerDestination, Distribution: 34},
				{Destination: BurnedDestination, Distribution: 33},
				{Destination: FoundationDestination, Distribution: 33},
			},
			inputs: []int64{10, 330},
			output: [][]int64{{4, 3, 3}, {114, 108, 108}},
		},
		{
			payouts: []Payout{
				{Destination: FarmerDestination, Distribution: 40},
				{Destination: BurnedDestination, Distribution: 40},
				{Destination: FoundationDestination, Distribution: 20},
			},
			inputs: []int64{10, 330},
			output: [][]int64{{4, 4, 2}, {132, 132, 66}},
		},
		{
			payouts: []Payout{
				{Destination: FarmerDestination, Distribution: 0},
				{Destination: BurnedDestination, Distribution: 73},
				{Destination: FoundationDestination, Distribution: 27},
			},
			inputs: []int64{10, 330},
			output: [][]int64{{0, 8, 2}, {0, 241, 89}},
		},
	}
	for _, pd := range pds {
		assert.NoError(t, pd.Valid())
	}

	w, err := stellar.New("", stellar.NetworkTest, nil)
	assert.NoError(t, err)

	e := NewStellar(w, nil, "", gridnetworks.GridNetworkMainnet)

	for i, tc := range payouts {
		for j, in := range tc.inputs {
			expected := tc.output[j]
			t.Run(fmt.Sprint(i, "_", j), func(t *testing.T) {
				amounts := e.splitPayout(xdr.Int64(in), tc.payouts)
				for x, a := range amounts {
					assert.Equal(t, expected[x], a)
				}
			})
		}
	}
}
