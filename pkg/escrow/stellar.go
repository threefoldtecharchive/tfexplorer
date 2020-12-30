package escrow

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/stellar/go/xdr"
	gdirectory "github.com/threefoldtech/tfexplorer/models/generated/directory"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	capacitytypes "github.com/threefoldtech/tfexplorer/pkg/capacity/types"
	"github.com/threefoldtech/tfexplorer/pkg/directory"
	directorytypes "github.com/threefoldtech/tfexplorer/pkg/directory/types"
	"github.com/threefoldtech/tfexplorer/pkg/escrow/types"
	"github.com/threefoldtech/tfexplorer/pkg/stellar"
	workloadtypes "github.com/threefoldtech/tfexplorer/pkg/workloads/types"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	// Stellar service manages a dedicate wallet for payments for reservations.
	Stellar struct {
		foundationAddress string
		wallet            *stellar.Wallet
		db                *mongo.Database

		reservationChannel         chan reservationRegisterJob
		capacityReservationChannel chan capacityReservationRegisterJob
		deployedChannel            chan schema.ID
		cancelledChannel           chan schema.ID

		paidCapacityInfoChannel chan schema.ID

		nodeAPI    NodeAPI
		gatewayAPI GatewayAPI
		farmAPI    FarmAPI

		ctx context.Context
	}

	// NodeAPI operations on node database
	NodeAPI interface {
		// Get a node from the database using its ID
		Get(ctx context.Context, db *mongo.Database, id string, proofs bool) (directorytypes.Node, error)
	}

	// GatewayAPI operations for gateway database
	GatewayAPI interface {
		// Get a gateway from the database using its ID
		Get(ctx context.Context, db *mongo.Database, id string) (directorytypes.Gateway, error)
	}

	// FarmAPI operations on farm database
	FarmAPI interface {
		// GetByID get a farm from the database using its ID
		GetByID(ctx context.Context, db *mongo.Database, id int64) (directorytypes.Farm, error)
	}

	reservationRegisterJob struct {
		reservation            workloads.Reservation
		supportedCurrencyCodes []string
		responseChan           chan reservationRegisterJobResponse
	}

	reservationRegisterJobResponse struct {
		data types.CustomerEscrowInformation
		err  error
	}

	capacityReservationRegisterJob struct {
		reservation            capacitytypes.Reservation
		supportedCurrencyCodes []string
		responseChan           chan capacityReservationRegisterJobResponse
	}

	capacityReservationRegisterJobResponse struct {
		data types.CustomerCapacityEscrowInformation
		err  error
	}
)

const (
	// interval between every check of active escrow accounts
	balanceCheckInterval = time.Minute * 1

	// maximum time for a capacity reservation
	capacityReservationTimeout = time.Hour * 1
)

const (
	// amount of digits of precision a calculated reservation cost has, at worst
	costPrecision = 6
)

var (
	// ErrNoCurrencySupported indicates a reservation was offered but none of the currencies
	// the farmer wants to pay in are currently supported
	ErrNoCurrencySupported = errors.New("none of the offered currencies are currently supported")
	// ErrNoCurrencyShared indicates that none of the currencies offered in the reservation
	// is supported by all farmers used
	ErrNoCurrencyShared = errors.New("none of the provided currencies is supported by all farmers")
)

var (
	totalReservationsProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "escrow",
		Name:      "total_reservations_processed",
		Help:      "The total number of reservations processed",
	})
	totalReservationsExpires = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "escrow",
		Name:      "total_reservations_expired",
		Help:      "The total number of reservations expired",
	})
	totalStellarTransactions = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "escrow",
		Name:      "total_transactions",
		Help:      "The total number of stellar transactions made",
	})
	totalNewEscrows = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "escrow",
		Name:      "new_escrows",
		Help:      "The total number of new escrows",
	})
	totalActiveEscrows = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "escrow",
		Name:      "active_escrows",
		Help:      "The total number of escrows active",
	})
	totalEscrowsPaid = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "escrow",
		Name:      "paid_escrows",
		Help:      "The total number of escrows paid",
	})
)

func init() {
	prometheus.MustRegister(totalReservationsProcessed)
	prometheus.MustRegister(totalReservationsExpires)
	prometheus.MustRegister(totalStellarTransactions)
	prometheus.MustRegister(totalNewEscrows)
	prometheus.MustRegister(totalActiveEscrows)
	prometheus.MustRegister(totalEscrowsPaid)
}

// NewStellar creates a new escrow object and fetches all addresses for the escrow wallet
func NewStellar(wallet *stellar.Wallet, db *mongo.Database, foundationAddress string) *Stellar {
	addr := foundationAddress
	if addr == "" {
		addr = wallet.PublicAddress()
	}

	return &Stellar{
		wallet:             wallet,
		db:                 db,
		foundationAddress:  addr,
		nodeAPI:            &directory.NodeAPI{},
		gatewayAPI:         &directory.GatewayAPI{},
		farmAPI:            &directory.FarmAPI{},
		reservationChannel: make(chan reservationRegisterJob),
		deployedChannel:    make(chan schema.ID),
		cancelledChannel:   make(chan schema.ID),
		// paidCapacityInfoChannel is buffered since it is used to communicate
		// with other workers, which might also try to communicate with this
		// worker
		paidCapacityInfoChannel:    make(chan schema.ID, 100),
		capacityReservationChannel: make(chan capacityReservationRegisterJob),
	}
}

// GetNetwork gets the explorer network
func (e *Stellar) GetNetwork() string {
	return e.wallet.GetNetwork()
}

// Run the escrow until the context is done
func (e *Stellar) Run(ctx context.Context) error {
	ticker := time.NewTicker(balanceCheckInterval)
	defer ticker.Stop()

	e.ctx = ctx

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("escrow context done, exiting")
			return nil

		case <-ticker.C:
			totalActiveEscrows.Set(0)
			log.Info().Msg("scanning active escrow accounts balance")
			if err := e.checkReservations(); err != nil {
				log.Error().Err(err).Msgf("failed to check reservations")
			}

			log.Info().Msg("scanning active capacity escrow accounts balance")
			if err := e.checkCapacityReservations(); err != nil {
				log.Error().Err(err).Msgf("failed to check capacity reservations")
			}

			log.Info().Msg("scanning for expired escrows")
			if err := e.refundExpiredReservations(); err != nil {
				log.Error().Err(err).Msgf("failed to refund expired reservations")
			}

			log.Info().Msg("scanning for expired capacity escrows")
			if err := e.refundExpiredCapacityReservations(); err != nil {
				log.Error().Err(err).Msgf("failed to refund expired capacity reservations")
			}

		case job := <-e.reservationChannel:
			log.Info().Int64("reservation_id", int64(job.reservation.ID)).Msg("processing new reservation escrow for reservation")
			details, err := e.processReservation(job.reservation, job.supportedCurrencyCodes)
			if err != nil {
				log.Error().
					Err(err).
					Int64("reservation_id", int64(job.reservation.ID)).
					Msgf("failed to check reservations")
			}
			job.responseChan <- reservationRegisterJobResponse{
				err:  err,
				data: details,
			}

		case job := <-e.capacityReservationChannel:
			log.Info().Int64("reservation_id", int64(job.reservation.ID)).Msg("processing new reservation escrow for reservation")
			details, err := e.processCapacityReservation(job.reservation, job.supportedCurrencyCodes)
			if err != nil {
				log.Error().
					Err(err).
					Int64("reservation_id", int64(job.reservation.ID)).
					Msgf("failed to check reservations")
			}
			job.responseChan <- capacityReservationRegisterJobResponse{
				err:  err,
				data: details,
			}
			totalReservationsProcessed.Inc()

		case id := <-e.deployedChannel:
			log.Info().Int64("reservation_id", int64(id)).Msg("trying to pay farmer for deployed reservation")
			if err := e.payoutFarmers(id); err != nil {
				log.Error().
					Err(err).
					Int64("reservation_id", int64(id)).
					Msgf("failed to payout farmers")
			}

		case id := <-e.cancelledChannel:
			log.Info().Int64("reservation_id", int64(id)).Msg("trying to refund clients for canceled reservation")
			if err := e.refundClients(id); err != nil {
				log.Error().
					Err(err).
					Int64("reservation_id", int64(id)).
					Msgf("could not refund clients")
			}
			totalReservationsExpires.Inc()
		}
	}
}

func (e *Stellar) refundExpiredReservations() error {
	// load expired escrows
	reservationEscrows, err := types.GetAllExpiredReservationPaymentInfos(e.ctx, e.db)
	if err != nil {
		return errors.Wrap(err, "failed to load active reservations from escrow")
	}
	for _, escrowInfo := range reservationEscrows {
		log.Info().Int64("id", int64(escrowInfo.ReservationID)).Msg("expired escrow")

		if err := e.refundEscrow(escrowInfo); err != nil {
			log.Error().Err(err).Msgf("failed to refund reservation escrow")
			continue
		}

		escrowInfo.Canceled = true
		if err = types.ReservationPaymentInfoUpdate(e.ctx, e.db, escrowInfo); err != nil {
			log.Error().Err(err).Msgf("failed to mark expired reservation escrow info as cancelled")
		}
	}
	return nil
}

func (e *Stellar) refundExpiredCapacityReservations() error {
	// load expired escrows
	reservationEscrows, err := types.GetAllExpiredCapacityReservationPaymentInfos(e.ctx, e.db)
	if err != nil {
		return errors.Wrap(err, "failed to load active reservations from escrow")
	}
	for _, escrowInfo := range reservationEscrows {
		log.Info().Int64("id", int64(escrowInfo.ReservationID)).Msg("expired escrow")

		if err := e.refundCapacityEscrow(escrowInfo); err != nil {
			log.Error().Err(err).Msgf("failed to refund reservation escrow")
			continue
		}

	}
	return nil
}

// checkReservations checks all the active reservations and marks those who are funded.
// if a reservation is funded then it will mark this reservation as to DEPLOY.
// if its underfunded it will throw an error.
func (e *Stellar) checkReservations() error {
	// load active escrows
	reservationEscrows, err := types.GetAllActiveReservationPaymentInfos(e.ctx, e.db)
	if err != nil {
		return errors.Wrap(err, "failed to load active reservations from escrow")
	}

	for _, escrowInfo := range reservationEscrows {
		if err := e.checkReservationPaid(escrowInfo); err != nil {
			log.Error().
				Str("address", escrowInfo.Address).
				Int64("reservation_id", int64(escrowInfo.ReservationID)).
				Err(err).
				Msg("failed to check reservation escrow funding status")
			continue
		}
		totalActiveEscrows.Inc()
	}
	return nil
}

// checkCapacityReservations checks all the active capacity reservations and marks
// those who are funded.
func (e *Stellar) checkCapacityReservations() error {
	// load active escrows
	reservationEscrows, err := types.GetAllActiveCapacityReservationPaymentInfos(e.ctx, e.db)
	if err != nil {
		return errors.Wrap(err, "failed to load active capacity reservations from escrow")
	}

	for _, escrowInfo := range reservationEscrows {
		if err := e.checkCapacityReservationPaid(escrowInfo); err != nil {
			log.Error().
				Str("address", escrowInfo.Address).
				Int64("reservation_id", int64(escrowInfo.ReservationID)).
				Err(err).
				Msg("failed to check reservation escrow funding status")
			continue
		}
		totalActiveEscrows.Inc()
		totalNewEscrows.Inc()
	}
	return nil
}

// CheckReservationPaid verifies if an escrow account received sufficient balance
// to pay for a reservation. If this is the case, the reservation will be moved
// to the deploy state, and the escrow state will be updated to indicate that this
// escrow has indeed been paid for this reservation, so it is not checked anymore
// in the future.
func (e *Stellar) checkReservationPaid(escrowInfo types.ReservationPaymentInformation) error {
	slog := log.With().
		Str("address", escrowInfo.Address).
		Int64("reservation_id", int64(escrowInfo.ReservationID)).
		Logger()

	// calculate total amount needed for reservation
	requiredValue := xdr.Int64(0)
	for _, escrowAccount := range escrowInfo.Infos {
		requiredValue += escrowAccount.TotalAmount
	}

	balance, _, err := e.wallet.GetBalance(escrowInfo.Address, reservationMemo(escrowInfo.ReservationID), escrowInfo.Asset)
	if err != nil {
		return errors.Wrap(err, "failed to verify escrow account balance")
	}

	if balance < requiredValue {
		slog.Debug().Msgf("required balance %d not reached yet (%d)", requiredValue, balance)
		return nil
	}

	slog.Debug().Msgf("required balance %d funded (%d), continue reservation", requiredValue, balance)

	reservation, err := workloadtypes.ReservationFilter{}.WithID(escrowInfo.ReservationID).Get(e.ctx, e.db)
	if err != nil {
		return errors.Wrap(err, "failed to load reservation")
	}

	pl, err := workloadtypes.NewPipeline(reservation)
	if err != nil {
		return errors.Wrap(err, "failed to process reservation pipeline")
	}

	reservation, _ = pl.Next()
	if !reservation.IsAny(workloadtypes.Pay) {
		// Do not continue, but also take no action to drive the reservation
		// as much as possible from the main explorer part.
		slog.Warn().Msg("reservation is paid, but no longer in pay state")
		// We warn because this is an unusual state to be in, yet there are
		// situations where this could happen. For example, we load the escrow,
		// the explorer then invalidates the actual reservation (e.g. user cancels),
		// we then load the updated reservation, which is no longer in pay state,
		// but the explorer is still cancelling the escrow, so we get here. As stated
		// above, we drive the escrow as much as possible from the workloads, with the
		// timeouts coming from the escrow itself, so this situation should always
		// resole itself. If we notice this log is coming back periodically, it thus means
		// there is a bug somewhere else in the code.
		// As a result, this state is therefore not considered an error.
		return nil
	}

	slog.Info().Msg("all farmer are paid, trying to move to deploy state")

	if err := workloadtypes.ReservationToDeploy(e.ctx, e.db, &reservation); err != nil {
		return errors.Wrap(err, "failed to schedule the reservation to deploy")
	}

	escrowInfo.Paid = true
	if err = types.ReservationPaymentInfoUpdate(e.ctx, e.db, escrowInfo); err != nil {
		return errors.Wrap(err, "failed to mark reservation escrow info as paid")
	}
	totalEscrowsPaid.Inc()

	slog.Debug().Msg("escrow marked as paid")

	return nil
}

// CheckCapacityReservationPaid verifies if an escrow account received sufficient balance
// to pay for a capacity reservation. If this is the case, the escrow state will
// be updated to the paid state, and the information regarding the capacity
// will be made available on the PaidCapacity channel.
// This also pais the farmer.
func (e *Stellar) checkCapacityReservationPaid(escrowInfo types.CapacityReservationPaymentInformation) error {
	slog := log.With().
		Str("address", escrowInfo.Address).
		Int64("reservation_id", int64(escrowInfo.ReservationID)).
		Logger()

	// calculate total amount needed for reservation
	requiredValue := escrowInfo.Amount

	balance, _, err := e.wallet.GetBalance(escrowInfo.Address, capacityReservationMemo(escrowInfo.ReservationID), escrowInfo.Asset)
	if err != nil {
		return errors.Wrap(err, "failed to verify escrow account balance")
	}

	if balance < requiredValue {
		slog.Debug().Msgf("required balance %d not reached yet (%d)", requiredValue, balance)
		return nil
	}

	slog.Debug().Msgf("required balance %d funded (%d), continue reservation", requiredValue, balance)

	escrowInfo.Paid = true
	if err = types.CapacityReservationPaymentInfoUpdate(e.ctx, e.db, escrowInfo); err != nil {
		return errors.Wrap(err, "failed to mark reservation escrow info as paid")
	}

	if err = e.payoutFarmersCap(escrowInfo); err != nil {
		slog.Debug().Msgf("farmer payout for capacity reservation %d failed, refund client", escrowInfo.ReservationID)
		if err2 := e.refundCapacityEscrow(escrowInfo); err2 != nil {
			// just log the error and return the main error
			log.Error().Err(err2).Msg("could not refund client")
		}
		return errors.Wrap(err, "failed to mark reservation escrow info as paid")
	}

	slog.Debug().Msg("escrow marked as paid")
	e.paidCapacityInfoChannel <- escrowInfo.ReservationID
	return nil
}

// processReservation processes a single reservation
// calculates resources and their costs
func (e *Stellar) processReservation(reservation workloads.Reservation, offeredCurrencyCodes []string) (types.CustomerEscrowInformation, error) {
	return types.CustomerEscrowInformation{}, fmt.Errorf("new reservations are not accepted, please use capacity pools and workloads API instead")
}

// processCapacityReservation processes a single reservation
// calculates resources and their costs
func (e *Stellar) processCapacityReservation(reservation capacitytypes.Reservation, offeredCurrencyCodes []string) (types.CustomerCapacityEscrowInformation, error) {
	var customerInfo types.CustomerCapacityEscrowInformation

	// filter out unsupported currencies
	currencies := []stellar.Asset{}
	for _, offeredCurrency := range offeredCurrencyCodes {
		asset, err := e.wallet.AssetFromCode(offeredCurrency)
		if err != nil {
			if err == stellar.ErrAssetCodeNotSupported {
				continue
			}
			return customerInfo, err
		}
		// Sanity check
		if _, exists := assetDistributions[asset]; !exists {
			// no payout distribution info set, log error and treat as if the asset
			// is not supported
			log.Error().Msgf("asset %s supported by wallet but no payout distribution found in escrow", asset)
			continue
		}
		currencies = append(currencies, asset)
	}

	if len(currencies) == 0 {
		return customerInfo, ErrNoCurrencySupported
	}

	// Get the farm ID. For now a pool must span only a single farm, therefore
	// all nodeID's must belong to the same farm. This must have been checked
	// already in a higher lvl handler.
	farmIDs := make([]int64, 1)
	node, err := e.nodeAPI.Get(e.ctx, e.db, reservation.DataReservation.NodeIDs[0], false)
	if err != nil {
		// TODO: proper abstraction
		if errors.Is(err, mongo.ErrNoDocuments) {
			// check if the ID is actually a gateway
			gw, err := e.gatewayAPI.Get(e.ctx, e.db, reservation.DataReservation.NodeIDs[0])
			if err != nil {
				return customerInfo, errors.Wrap(err, "could not get gateway")
			}
			farmIDs[0] = gw.FarmId
		} else {
			return customerInfo, errors.Wrap(err, "could not get node")
		}
	} else {
		farmIDs[0] = node.FarmId
	}

	// check which currencies are accepted by all farmers
	// the farm ids have conveniently been provided when checking the used rsu
	var asset stellar.Asset
	for _, currency := range currencies {
		// if the farmer does not receive anything in the first place, they always
		// all agree on this currency
		if assetDistributions[currency].farmer == 0 {
			asset = currency
			break
		}
		// check if all used farms have an address for this asset set up
		supported, err := e.checkAssetSupport(farmIDs, currency)
		if err != nil {
			return customerInfo, errors.Wrap(err, "could not verify asset support")
		}
		if supported {
			asset = currency
			break
		}
	}

	if asset == "" {
		return customerInfo, ErrNoCurrencyShared
	}

	address, err := e.createOrLoadAccount(reservation.CustomerTid)
	if err != nil {
		return customerInfo, errors.Wrap(err, "failed to get escrow address for customer")
	}

	amount, err := e.calculateCapacityReservationCost(reservation.DataReservation.CUs, reservation.DataReservation.SUs, reservation.DataReservation.IPv4Us, node.FarmId)
	if err != nil {
		return customerInfo, errors.Wrap(err, "failed to calculate capacity cost")
	}

	reservationPaymentInfo := types.CapacityReservationPaymentInformation{
		ReservationID: reservation.ID,
		Address:       address,
		Expiration:    schema.Date{Time: time.Now().Add(capacityReservationTimeout)},
		Asset:         asset,
		Amount:        amount,
		Paid:          false,
		Released:      false,
		Canceled:      false,
		FarmerID:      schema.ID(node.FarmId),
	}
	if amount == 0 {
		// mark this reservation as fully processed already
		reservationPaymentInfo.Paid = true
		reservationPaymentInfo.Released = true
		log.Debug().Int64("id", int64(reservation.ID)).Msg("0 value reservation, mark as processed")
	}
	err = types.CapacityReservationPaymentInfoCreate(e.ctx, e.db, reservationPaymentInfo)
	if err != nil {
		return customerInfo, errors.Wrap(err, "failed to create reservation payment information")
	}

	if amount == 0 {
		// Now that the info is successfully saved, notify that it has been paid
		log.Debug().Int64("id", int64(reservation.ID)).Msg("pushing reservation id on paid reservations channel")
		e.paidCapacityInfoChannel <- reservation.ID
	}
	log.Info().Int64("id", int64(reservation.ID)).Msg("processed reservation and created payment information")
	customerInfo.Address = address
	customerInfo.Asset = asset
	customerInfo.Amount = amount
	return customerInfo, nil
}

// refundClients refunds clients if the reservation is cancelled
func (e *Stellar) refundClients(id schema.ID) error {
	rpi, err := types.ReservationPaymentInfoGet(e.ctx, e.db, id)
	if err != nil {
		return errors.Wrap(err, "failed to get reservation escrow info")
	}
	if rpi.Released || rpi.Canceled {
		// already paid
		return nil
	}
	if err := e.refundEscrow(rpi); err != nil {
		log.Error().Err(err).Msg("failed to refund escrow")
		return errors.Wrap(err, "could not refund escrow")
	}

	rpi.Canceled = true
	if err = types.ReservationPaymentInfoUpdate(e.ctx, e.db, rpi); err != nil {
		return errors.Wrapf(err, "could not mark escrows for %d as canceled", rpi.ReservationID)
	}
	log.Debug().Int64("id", int64(rpi.ReservationID)).Msg("refunded clients for reservation")
	return nil
}

// payoutFarmers pays out the farmer for a processed reservation
func (e *Stellar) payoutFarmers(id schema.ID) error {
	rpi, err := types.ReservationPaymentInfoGet(e.ctx, e.db, id)
	if err != nil {
		if errors.Is(err, types.ErrEscrowNotFound) {
			// reservation deployed for which there is no escrow, this is possible
			// for a reservation attached to capacity pools. Don't worry about it.
			return nil
		}
		return errors.Wrap(err, "failed to get reservation escrow info")
	}
	if rpi.Released || rpi.Canceled {
		// already paid
		return nil
	}

	paymentDistribution, exists := assetDistributions[rpi.Asset]
	if !exists {
		return fmt.Errorf("no payment distribution found for asset %s", rpi.Asset)
	}

	// keep track of total amount to burn and to send to foundation
	var toBurn, toFoundation xdr.Int64

	// collect the farmer addresses and amount they should receive, we already
	// have sufficient balance on the escrow to cover this
	paymentInfo := make([]stellar.PayoutInfo, 0, len(rpi.Infos))

	for _, escrowDetails := range rpi.Infos {
		farmerAmount, burnAmount, foundationAmount := e.splitPayout(escrowDetails.TotalAmount, paymentDistribution)
		toBurn += burnAmount
		toFoundation += foundationAmount

		if farmerAmount > 0 {
			// in case of an error in this flow we continue, so we try to pay as much
			// farmers as possible even if one fails
			farm, err := e.farmAPI.GetByID(e.ctx, e.db, int64(escrowDetails.FarmerID))
			if err != nil {
				log.Error().Msgf("failed to load farm info: %s", err)
				continue
			}

			destination, err := addressByAsset(farm.WalletAddresses, rpi.Asset)
			if err != nil {
				// FIXME: this is probably not ok, what do we do in this case ?
				log.Error().Err(err).Msgf("failed to find address for %s for farmer %d", rpi.Asset.Code(), farm.ID)
				continue
			}

			// farmerAmount can't be pooled so add an info immediately
			paymentInfo = append(paymentInfo,
				stellar.PayoutInfo{
					Address: destination,
					Amount:  farmerAmount,
				},
			)
		}
	}

	// a burn is a transfer of tokens back to the issuer
	if toBurn > 0 {
		paymentInfo = append(paymentInfo,
			stellar.PayoutInfo{
				Address: rpi.Asset.Issuer(),
				Amount:  toBurn,
			})
	}

	// ship remainder to the foundation
	if toFoundation > 0 {
		paymentInfo = append(paymentInfo,
			stellar.PayoutInfo{
				Address: e.foundationAddress,
				Amount:  toFoundation,
			})
	}

	addressInfo, err := types.CustomerAddressByAddress(e.ctx, e.db, rpi.Address)
	if err != nil {
		log.Error().Msgf("failed to load escrow address info: %s", err)
		return errors.Wrap(err, "could not load escrow address info")
	}
	if err = e.wallet.PayoutFarmers(addressInfo.Secret, paymentInfo, reservationMemo(id), rpi.Asset); err != nil {
		log.Error().Msgf("failed to pay farmer: %s for reservation %d", err, id)
		return errors.Wrap(err, "could not pay farmer")
	}
	totalStellarTransactions.Inc()

	// now refund any possible overpayment
	if err = e.wallet.Refund(addressInfo.Secret, reservationMemo(id), rpi.Asset); err != nil {
		log.Error().Msgf("failed to refund overpayment farmer: %s", err)
		return errors.Wrap(err, "could not refund overpayment")
	}
	totalStellarTransactions.Inc()

	log.Info().
		Str("escrow address", rpi.Address).
		Int64("reservation id", int64(rpi.ReservationID)).
		Msgf("paid farmer")

	rpi.Released = true
	if err = types.ReservationPaymentInfoUpdate(e.ctx, e.db, rpi); err != nil {
		return errors.Wrapf(err, "could not mark escrows for %d as released", rpi.ReservationID)
	}
	return nil
}

// payoutFarmersCap pays out the farmer for a processed reservation
func (e *Stellar) payoutFarmersCap(rpi types.CapacityReservationPaymentInformation) error {
	if rpi.Released || rpi.Canceled {
		// already paid
		return nil
	}

	paymentDistribution, exists := assetDistributions[rpi.Asset]
	if !exists {
		return fmt.Errorf("no payment distribution found for asset %s", rpi.Asset)
	}

	// keep track of total amount to burn and to send to foundation
	var toBurn, toFoundation xdr.Int64

	farmerAmount, burnAmount, foundationAmount := e.splitPayout(rpi.Amount, paymentDistribution)
	toBurn += burnAmount
	toFoundation += foundationAmount

	paymentInfo := []stellar.PayoutInfo{}
	if farmerAmount > 0 {
		// in case of an error in this flow we continue, so we try to pay as much
		// farmers as possible even if one fails
		farm, err := e.farmAPI.GetByID(e.ctx, e.db, int64(rpi.FarmerID))
		if err != nil {
			return errors.Wrap(err, "failed to load farm info")
		}

		destination, err := addressByAsset(farm.WalletAddresses, rpi.Asset)
		if err != nil {
			return errors.Wrapf(err, "failed to find address for %s for farmer %d", rpi.Asset.Code(), farm.ID)
		}

		// farmerAmount can't be pooled so add an info immediately
		paymentInfo = append(paymentInfo,
			stellar.PayoutInfo{
				Address: destination,
				Amount:  farmerAmount,
			},
		)
	}

	// a burn is a transfer of tokens back to the issuer
	if toBurn > 0 {
		paymentInfo = append(paymentInfo,
			stellar.PayoutInfo{
				Address: rpi.Asset.Issuer(),
				Amount:  toBurn,
			})
	}

	// ship remainder to the foundation
	if toFoundation > 0 {
		paymentInfo = append(paymentInfo,
			stellar.PayoutInfo{
				Address: e.foundationAddress,
				Amount:  toFoundation,
			})
	}

	addressInfo, err := types.CustomerAddressByAddress(e.ctx, e.db, rpi.Address)
	if err != nil {
		log.Error().Msgf("failed to load escrow address info: %s", err)
		return errors.Wrap(err, "could not load escrow address info")
	}
	if err = e.wallet.PayoutFarmers(addressInfo.Secret, paymentInfo, capacityReservationMemo(rpi.ReservationID), rpi.Asset); err != nil {
		log.Error().Msgf("failed to pay farmer: %s for reservation %d", err, rpi.ReservationID)
		return errors.Wrap(err, "could not pay farmer")
	}
	totalStellarTransactions.Inc()

	// now refund any possible overpayment
	if err = e.wallet.Refund(addressInfo.Secret, capacityReservationMemo(rpi.ReservationID), rpi.Asset); err != nil {
		log.Error().Msgf("failed to refund overpayment farmer: %s", err)
		return errors.Wrap(err, "could not refund overpayment")
	}
	totalStellarTransactions.Inc()

	log.Info().
		Str("escrow address", rpi.Address).
		Int64("reservation id", int64(rpi.ReservationID)).
		Msgf("paid farmer")

	rpi.Released = true
	if err = types.CapacityReservationPaymentInfoUpdate(e.ctx, e.db, rpi); err != nil {
		return errors.Wrapf(err, "could not mark escrows for %d as released", rpi.ReservationID)
	}
	return nil
}

func (e *Stellar) refundEscrow(escrowInfo types.ReservationPaymentInformation) error {
	slog := log.With().
		Str("address", escrowInfo.Address).
		Int64("reservation_id", int64(escrowInfo.ReservationID)).
		Logger()

	slog.Info().Msgf("try to refund client for escrow")

	addressInfo, err := types.CustomerAddressByAddress(e.ctx, e.db, escrowInfo.Address)
	if err != nil {
		return errors.Wrap(err, "failed to load escrow info")
	}

	if err = e.wallet.Refund(addressInfo.Secret, reservationMemo(escrowInfo.ReservationID), escrowInfo.Asset); err != nil {
		return errors.Wrap(err, "failed to refund clients")
	}
	totalStellarTransactions.Inc()

	slog.Info().Msgf("refunded client for escrow")
	return nil
}

func (e *Stellar) refundCapacityEscrow(escrowInfo types.CapacityReservationPaymentInformation) error {
	slog := log.With().
		Str("address", escrowInfo.Address).
		Int64("reservation_id", int64(escrowInfo.ReservationID)).
		Logger()

	slog.Info().Msgf("try to refund client for escrow")

	addressInfo, err := types.CustomerAddressByAddress(e.ctx, e.db, escrowInfo.Address)
	if err != nil {
		return errors.Wrap(err, "failed to load escrow info")
	}

	if err = e.wallet.Refund(addressInfo.Secret, capacityReservationMemo(escrowInfo.ReservationID), escrowInfo.Asset); err != nil {
		return errors.Wrap(err, "failed to refund clients")
	}
	totalStellarTransactions.Inc()

	escrowInfo.Canceled = true
	if err = types.CapacityReservationPaymentInfoUpdate(e.ctx, e.db, escrowInfo); err != nil {
		return errors.Wrap(err, "failed to mark expired reservation escrow info as cancelled")
	}

	slog.Info().Msgf("refunded client for escrow")
	return nil
}

// RegisterReservation registers a workload reservation
func (e *Stellar) RegisterReservation(reservation workloads.Reservation, supportedCurrencies []string) (types.CustomerEscrowInformation, error) {
	job := reservationRegisterJob{
		reservation:            reservation,
		supportedCurrencyCodes: supportedCurrencies,
		responseChan:           make(chan reservationRegisterJobResponse),
	}
	e.reservationChannel <- job

	response := <-job.responseChan

	return response.data, response.err
}

// CapacityReservation implements Escrow
func (e *Stellar) CapacityReservation(reservation capacitytypes.Reservation, supportedCurrencies []string) (types.CustomerCapacityEscrowInformation, error) {
	job := capacityReservationRegisterJob{
		reservation:            reservation,
		supportedCurrencyCodes: supportedCurrencies,
		responseChan:           make(chan capacityReservationRegisterJobResponse),
	}
	e.capacityReservationChannel <- job

	response := <-job.responseChan

	return response.data, response.err
}

// PaidCapacity implements Escrow
func (e *Stellar) PaidCapacity() <-chan schema.ID {
	return e.paidCapacityInfoChannel
}

// ReservationDeployed informs the escrow that a reservation has been successfully
// deployed, so the escrow can release the funds to the farmer (and refund any excess)
func (e *Stellar) ReservationDeployed(reservationID schema.ID) {
	e.deployedChannel <- reservationID
}

// ReservationCanceled informs the escrow of a canceled reservation so it can refund
// the user
func (e *Stellar) ReservationCanceled(reservationID schema.ID) {
	e.cancelledChannel <- reservationID
}

// createOrLoadAccount creates or loads account based on  customer id
func (e *Stellar) createOrLoadAccount(customerTID int64) (string, error) {
	res, err := types.CustomerAddressGet(context.Background(), e.db, customerTID)
	if err != nil {
		if err == types.ErrAddressNotFound {
			seed, address, err := e.wallet.CreateAccount()
			if err != nil {
				return "", errors.Wrapf(err, "failed to create a new account for customer %d", customerTID)
			}
			totalStellarTransactions.Inc()

			err = types.CustomerAddressCreate(context.Background(), e.db, types.CustomerAddress{
				CustomerTID: customerTID,
				Address:     address,
				Secret:      seed,
			})
			if err != nil {
				return "", errors.Wrapf(err, "failed to save a new account for customer %d", customerTID)
			}
			log.Debug().
				Int64("customer", int64(customerTID)).
				Str("address", address).
				Msgf("created new escrow address for customer")

			return address, nil
		}
		return "", errors.Wrap(err, "failed to get customer address")
	}
	log.Debug().
		Int64("customer", int64(customerTID)).
		Str("address", res.Address).
		Msgf("escrow address found for customer")

	return res.Address, nil
}

// splitPayout to a farmer in the amount the farmer receives, the amount to be burned,
// and the amount the foundation receives
func (e *Stellar) splitPayout(totalAmount xdr.Int64, distribution payoutDistribution) (xdr.Int64, xdr.Int64, xdr.Int64) {
	// we can't just use big.Float for this calculation, since we need to verify
	// the rounding afterwards

	// calculate missing precision digits, to perform percentage division without
	// floating point operations
	requiredPrecision := 2 + costPrecision
	missingPrecision := requiredPrecision - e.wallet.PrecisionDigits()

	multiplier := int64(1)
	if missingPrecision > 0 {
		multiplier = int64(math.Pow10(missingPrecision))
	}

	amount := int64(totalAmount) * multiplier

	baseAmount := amount / 100
	farmerAmount := baseAmount * int64(distribution.farmer)
	burnAmount := baseAmount * int64(distribution.burned)
	foundationAmount := baseAmount * int64(distribution.foundation)

	// collect parts which will be missing in division, if any
	var change int64
	change += farmerAmount % multiplier
	change += burnAmount % multiplier
	change += foundationAmount % multiplier

	// change is now necessarily a multiple of multiplier
	change /= multiplier
	// we tracked all change which would be removed by the following integer
	// devisions
	farmerAmount /= multiplier
	burnAmount /= multiplier
	foundationAmount /= multiplier

	// give change to whichever gets funds anyway, in the following order:
	//  - farmer
	//  - burned
	//  - foundation
	if farmerAmount != 0 {
		farmerAmount += change
	} else if burnAmount != 0 {
		burnAmount += change
	} else if foundationAmount != 0 {
		foundationAmount += change
	}

	return xdr.Int64(farmerAmount), xdr.Int64(burnAmount), xdr.Int64(foundationAmount)
}

// checkAssetSupport for all unique farms in the reservation
func (e *Stellar) checkAssetSupport(farmIDs []int64, asset stellar.Asset) (bool, error) {
	for _, id := range farmIDs {
		farm, err := e.farmAPI.GetByID(e.ctx, e.db, id)
		if err != nil {
			return false, errors.Wrap(err, "could not load farm")
		}
		if _, err := addressByAsset(farm.WalletAddresses, asset); err != nil {
			// this only errors if the asset is not present
			return false, nil
		}
	}
	return true, nil
}

func addressByAsset(addrs []gdirectory.WalletAddress, asset stellar.Asset) (string, error) {
	for _, a := range addrs {
		if a.Asset == asset.Code() && a.Address != "" {
			return a.Address, nil
		}
	}
	return "", fmt.Errorf("not address found for asset %s", asset)
}

func reservationMemo(id schema.ID) string {
	return fmt.Sprintf("%d", id)
}

func capacityReservationMemo(id schema.ID) string {
	return fmt.Sprintf("p-%d", id)
}
