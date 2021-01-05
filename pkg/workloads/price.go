package workloads

import (
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfexplorer/mw"
	"github.com/threefoldtech/tfexplorer/pkg/escrow"
)

var (
	pricesOnce sync.Once
	prices     struct {
		CuPriceDollarMonth   float64
		SuPriceDollarMonth   float64
		TftPriceMill         float64
		IP4uPriceDollarMonth float64
	}
)

func (a *API) getPrices(r *http.Request) (interface{}, mw.Response) {
	pricesOnce.Do(func() {
		div, err := a.network.Divisor()
		if err != nil {
			log.Error().Err(err).Msg("failed to get network divisor")
			div = 1
		}

		divisor := float64(div)

		prices.CuPriceDollarMonth = float64(escrow.CuPriceDollarMonth) / divisor
		prices.SuPriceDollarMonth = float64(escrow.SuPriceDollarMonth) / divisor
		prices.TftPriceMill = escrow.TftPriceMill
		prices.IP4uPriceDollarMonth = float64(escrow.IP4uPriceDollarMonth) / divisor
	})

	return prices, nil
}
