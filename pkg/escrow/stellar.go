package escrow

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/go/xdr"
	gdirectory "github.com/threefoldtech/tfexplorer/models/generated/directory"
	capacitytypes "github.com/threefoldtech/tfexplorer/pkg/capacity/types"
	"github.com/threefoldtech/tfexplorer/pkg/directory"
	directorytypes "github.com/threefoldtech/tfexplorer/pkg/directory/types"
	"github.com/threefoldtech/tfexplorer/pkg/escrow/types"
	"github.com/threefoldtech/tfexplorer/pkg/gridnetworks"
	phonebooktypes "github.com/threefoldtech/tfexplorer/pkg/phonebook/types"
	"github.com/threefoldtech/tfexplorer/pkg/stellar"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	// Stellar service manages a dedicate wallet for payments for reservations.
	Stellar struct {
		foundationAddress string
		wallet            stellar.Wallet
		db                *mongo.Database
		gridNetwork       gridnetworks.GridNetwork

		capacityReservationChannel chan capacityReservationRegisterJob

		paidCapacityInfoChannel chan schema.ID

		paymentsChannel chan stellar.PayoutJob

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
		GetFarmCustomPriceForThreebot(ctx context.Context, db *mongo.Database, farmID, threebotID int64) (directorytypes.FarmThreebotPrice, error)
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
	balanceCheckInterval = time.Second * 5

	// maximum time for a capacity reservation
	capacityReservationTimeout = time.Hour * 1

	// MaxSignaturesPerTx maximum number of signatures in a transaction
	MaxSignaturesPerTx = 20

	// MaxOperationsPerTx maximum number of operations in a transaction
	MaxOperationsPerTx = 100
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

func getBatchMemoTransactions(ctx context.Context, db *mongo.Database, memo string) (stellar.BatchTransactionsInfo, error) {
	memoInfos, err := types.CapacityMemoTextInfoGet(ctx, db, memo)
	res := stellar.BatchTransactionsInfo{
		Ops: make(map[string][]int),
	}
	if err != nil {
		return res, errors.Wrap(err, "couldn't get capacity memo text info")
	}
	for _, info := range memoInfos {
		res.Ops[info.TxSequence] = info.OperationIDs
	}
	return res, nil
}

// NewStellar creates a new escrow object and fetches all addresses for the escrow wallet
func NewStellar(wallet stellar.Wallet, db *mongo.Database, foundationAddress string, gridNetwork gridnetworks.GridNetwork) *Stellar {
	addr := foundationAddress
	if addr == "" {
		addr = wallet.PublicAddress()
	}

	return &Stellar{
		wallet:            wallet,
		db:                db,
		foundationAddress: addr,
		gridNetwork:       gridNetwork,
		nodeAPI:           &directory.NodeAPI{},
		gatewayAPI:        &directory.GatewayAPI{},
		farmAPI:           &directory.FarmAPI{},
		paymentsChannel:   make(chan stellar.PayoutJob, 100),
		// paidCapacityInfoChannel is buffered since it is used to communicate
		// with other workers, which might also try to communicate with this
		// worker
		paidCapacityInfoChannel:    make(chan schema.ID, 100),
		capacityReservationChannel: make(chan capacityReservationRegisterJob),
	}
}

// RepushPendingPayments at initialization push all unprocessed pending payments
func (e *Stellar) RepushPendingPayments() error {
	pendingCancellations, err := types.CapacityReservationPaymentInfoPendingCancellationGet(e.ctx, e.db)
	if err != nil {
		return errors.Wrap(err, "failed to get pending payment cancellations")
	}
	pendingPayments, err := types.CapacityReservationPaymentInfoPendingPayoutsGet(e.ctx, e.db)
	if err != nil {
		return errors.Wrap(err, "failed to get pending payments")
	}
	for _, escrowInfo := range pendingCancellations {
		err = e.refundCapacityEscrow(escrowInfo, escrowInfo.Cause)
		if err != nil {
			log.Error().Err(err).Int("id", int(escrowInfo.ReservationID)).Msg("failed to push refund to the queue")
		}
	}
	for _, escrowInfo := range pendingPayments {
		err = e.payoutFarmersCap(escrowInfo)
		if err != nil {
			log.Error().Err(err).Int("id", int(escrowInfo.ReservationID)).Msg("failed to push payment to the queue")
		}
	}
	return nil
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

			log.Info().Msg("scanning active capacity escrow accounts balance")
			if err := e.checkCapacityReservations(); err != nil {
				log.Error().Err(err).Msgf("failed to check capacity reservations")
			}

			log.Info().Msg("scanning for expired capacity escrows")
			if err := e.refundExpiredCapacityReservations(); err != nil {
				log.Error().Err(err).Msgf("failed to refund expired capacity reservations")
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
		}

	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (e *Stellar) processFailedPayments(herr *horizonclient.Error, jobs []stellar.PayoutJob, payments []txnbuild.Payment, jobMap map[int]int) error {
	badJob := make(map[int]bool)
	xdr, err := herr.EnvelopeXDR()
	if err != nil {
		log.Error().Err(err).Msg("couldn't extract xdr from failed transaction")
		xdr = ""
	}
	resString, err := herr.ResultString()
	if err != nil {
		log.Error().Err(err).Msg("couldn't extract result string from failed transaction")
		resString = ""
	}

	codes, reserr := herr.ResultCodes()
	transactionFailed := false
	var operationCodes []string
	if reserr != nil || len(codes.OperationCodes) == 0 {
		// repush all if the whole transaction failed
		log.Debug().Err(err).Msg("couldn't get result codes or no op codes")
		transactionFailed = true
	} else {
		operationCodes = codes.OperationCodes

	}
	for i, code := range operationCodes {
		if !transactionFailed && code != "op_success" {
			log.Error().
				Str("code", code).
				Str("source account", payments[i].SourceAccount.GetAccountID()).
				Str("destination account", payments[i].Destination).
				Str("amount", payments[i].Amount).
				Str("memo text", jobs[jobMap[i]].Memo).
				Str("code", code).
				Msg("operation failed")
			badJob[jobMap[i]] = true
		}
	}
	var curOpIdx int
	for i, j := range jobs {
		curOpIdx += len(j.Payments)
		// repush non-failed transactions in the queue
		escrowInfo, err := types.CapacityReservationPaymentInfoGet(e.ctx, e.db, j.ID)
		if err != nil {
			log.Error().Err(err).Msg("could not get reservation id for failed payment")
			continue
		}
		if !badJob[i] {
			e.paymentsChannel <- j
		} else if j.Retries > 1 {
			j.Retries--
			e.paymentsChannel <- j
		} else {
			// refund failed payments and store failed refunds in the db
			if !j.Refund {
				e.refundCapacityEscrow(escrowInfo, "farmer_payout")
			} else {
				fmt.Printf("code: %s, idx: %d, opcodes: %v", operationCodes[i], i, operationCodes)
				escrowInfo.Cause = fmt.Sprintf("Failed to refund user, original error: %s", escrowInfo.Cause)
				escrowInfo.CancellationPending = false
				escrowInfo.Canceled = true // we gave up
				failed := types.FailedPaymentInfo{
					ReservationID: j.ID,
					MemoText:      j.Memo,
					ErrorCodes:    operationCodes[curOpIdx-len(j.Payments) : curOpIdx],
					EnvelopeXDR:   xdr,
					ResultString:  resString,
				}
				err := types.FailedPaymentInfoInfoCreate(e.ctx, e.db, failed)
				if err != nil {
					log.Error().Err(err).Msg("failed to push the failed payment to the db")
				}
				err = types.CapacityReservationPaymentInfoUpdate(e.ctx, e.db, escrowInfo)
				if err != nil {
					log.Error().Err(err).Msg("failed to update the escrow info in the db")
				}
			}
		}
	}
	return nil
}

// PaymentsLoop the payment loop the context is done
func (e *Stellar) PaymentsLoop(ctx context.Context) error {
	for {
		var secrets []string
		var payments []txnbuild.Payment
		var jobs []stellar.PayoutJob
		jobMap := make(map[int]int)
		memoMap := make(map[string][]int)
		ready := false
		for !ready {
			select {
			case <-ctx.Done():
				log.Info().Msg("escrow context done, exiting")
				return nil

			case job := <-e.paymentsChannel:
				jobs = append(jobs, job)
				memoMap[job.Memo] = make([]int, 0)
				if !stringInSlice(job.SecretKey, secrets) {
					secrets = append(secrets, job.SecretKey)
				}
				for _, payment := range job.Payments {
					log.Debug().
						Str("amount", payment.Amount).
						Str("source", payment.SourceAccount.GetAccountID()).
						Str("destination", payment.Destination).
						Str("asset", job.Asset.String()).
						Msg("another payment received")
					memoMap[job.Memo] = append(memoMap[job.Memo], len(payments))
					jobMap[len(payments)] = len(jobs) - 1
					payments = append(payments, payment)
				}
				if len(payments) >= MaxOperationsPerTx*.8 || len(secrets) >= MaxSignaturesPerTx-1 {
					ready = true
				}
			default:
				if len(secrets) > 0 {
					ready = true
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}
		sequenctNumber, err := e.wallet.GetNextSequenceNumber()
		if err != nil {
			log.Error().Msgf("failed to get sequence number: %s", err)
		}
		for memo, ops := range memoMap {
			types.CapacityMemoTextInfoCreate(ctx, e.db, types.CapacityMemoTextInfo{
				MemoText:     memo,
				OperationIDs: ops,
				TxSequence:   sequenctNumber,
			})
		}
		err = e.wallet.ProcessPayoutBatches(payments, secrets)
		totalStellarTransactions.Inc()
		if err != nil {
			if err2, ok := err.(*horizonclient.Error); ok {
				err = e.processFailedPayments(err2, jobs, payments, jobMap)
			}
			log.Error().Msgf("failed to submit all payouts: %s", err)
			continue
		}
		log.Debug().Msg("End submitting batches")
		for id, job := range jobs {
			rpi, err := types.CapacityReservationPaymentInfoGet(ctx, e.db, jobs[id].ID)
			if err != nil {
				log.Error().Msgf("failed to get payment info by id: %s", err)
			}
			if !job.Refund {
				rpi.Released = true
				e.paidCapacityInfoChannel <- rpi.ReservationID
			} else {
				rpi.CancellationPending = false
				rpi.Canceled = true

			}
			if err = types.CapacityReservationPaymentInfoUpdate(e.ctx, e.db, rpi); err != nil {
				log.Error().Err(err).Msgf("could not mark escrows for %d as released", rpi.ReservationID)
			}
		}
	}
}

func (e *Stellar) refundExpiredCapacityReservations() error {
	// load expired escrows
	reservationEscrows, err := types.GetAllExpiredCapacityReservationPaymentInfos(e.ctx, e.db)
	if err != nil {
		return errors.Wrap(err, "failed to load active reservations from escrow")
	}
	for _, escrowInfo := range reservationEscrows {
		log.Info().Int64("id", int64(escrowInfo.ReservationID)).Msg("expired escrow")

		if err := e.refundCapacityEscrow(escrowInfo, "expired"); err != nil {
			log.Error().Err(err).Msgf("failed to refund reservation escrow")
			continue
		}

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
	memo := capacityReservationMemo(escrowInfo.ReservationID)
	batchTx, err := getBatchMemoTransactions(e.ctx, e.db, memo)
	if err != nil {
		log.Error().Err(err).Str("memo", memo).Msg("failed to get batch memo transactions")
	}
	balance, _, err := e.wallet.GetBalance(escrowInfo.Address, memo, escrowInfo.Asset, &batchTx)
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
		if err2 := e.refundCapacityEscrow(escrowInfo, err.Error()); err2 != nil {
			// just log the error and return the main error
			log.Error().Err(err2).Msg("could not refund client")
		}
		return errors.Wrap(err, "failed to mark reservation escrow info as paid")
	}

	slog.Debug().Msg("escrow marked as paid")
	return nil
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
		// we currently only support ONE asset
		if asset != stellar.TFTMainnet {
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
	var amount xdr.Int64
	whichThreebotID := reservation.CustomerTid
	poolID := reservation.ID
	if reservation.DataReservation.PoolID != 0 {
		poolID = schema.ID(reservation.DataReservation.PoolID)
	}

	pool, err := capacitytypes.GetPool(e.ctx, e.db, schema.ID(poolID))
	if err != nil {
		return types.CustomerCapacityEscrowInformation{}, err
	}
	if pool.SponsorTid != 0 {
		whichThreebotID = pool.SponsorTid
	}

	price, err := e.farmAPI.GetFarmCustomPriceForThreebot(e.ctx, e.db, farmIDs[0], whichThreebotID)
	// safe to ignore the error here, we already have a farm
	if err != nil {
		amount, err = e.calculateCapacityReservationCost(reservation.DataReservation.CUs, reservation.DataReservation.SUs, reservation.DataReservation.IPv4Us)
		if err != nil {
			return customerInfo, errors.Wrap(err, "failed to calculate capacity reservation cost")
		}
	} else {
		cuDollarPerMonth := price.CustomCloudUnitPrice.CU
		suDollarPerMonth := price.CustomCloudUnitPrice.SU
		ip4uDollarPerMonth := price.CustomCloudUnitPrice.IPv4U
		amount, err = e.calculateCustomCapacityReservationCost(reservation.DataReservation.CUs, reservation.DataReservation.SUs, reservation.DataReservation.IPv4Us, cuDollarPerMonth, suDollarPerMonth, ip4uDollarPerMonth)
		if err != nil {
			return customerInfo, errors.Wrap(err, "failed to calculate capacity reservation cost")
		}
	}

	reservationPaymentInfo := types.CapacityReservationPaymentInformation{
		ReservationID:       reservation.ID,
		Address:             address,
		Expiration:          schema.Date{Time: time.Now().Add(capacityReservationTimeout)},
		Asset:               asset,
		Amount:              amount,
		Paid:                false,
		Released:            false,
		Canceled:            false,
		CancellationPending: false,
		FarmerID:            schema.ID(farmIDs[0]),
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

func (e *Stellar) getPool(reservationID schema.ID) (capacitytypes.Pool, error) {
	reservation, err := capacitytypes.CapacityReservationGet(e.ctx, e.db, reservationID)
	if err != nil {
		return capacitytypes.Pool{}, err
	}

	poolID := reservation.ID
	if reservation.DataReservation.PoolID != 0 {
		poolID = schema.ID(reservation.DataReservation.PoolID)
	}

	return capacitytypes.GetPool(e.ctx, e.db, poolID)
}

// asset distribution based on issue https://github.com/threefoldtech/home/issues/1041
func (e *Stellar) getPayouts(rpi types.CapacityReservationPaymentInformation, farm directorytypes.Farm) ([]Payout, error) {
	var (
		distribution      PaymentDistribution
		farmerAddress     string
		salesAddress      string
		foundationAddress = e.foundationAddress
		wisdomAddress     = WisdomWallet
	)

	var err error
	farmerAddress, err = addressByAsset(farm.WalletAddresses, rpi.Asset)
	if err != nil {
		return nil, err
	}

	if !farm.IsGrid3Compliant {
		// grid 2
		distribution = AssetDistributions[DistributionV2]
	} else {
		// grid 3 default distribution
		distribution = AssetDistributions[DistributionV3]

		pool, err := e.getPool(rpi.ReservationID)
		if err != nil {
			return nil, err
		}

		// check if certified sales channel
		// this can be detected if the pool is sponsored
		if pool.SponsorTid != 0 {
			// sponsor channel.
			distribution = AssetDistributions[DistributionCertifiedSales]
			// fill in the address for the sales channel
			var f phonebooktypes.UserFilter
			f = f.WithID(schema.ID(pool.SponsorTid))
			sales, err := f.Get(e.ctx, e.db)
			if err != nil {
				return nil, err
			}

			salesAddress, err = addressByAsset(sales.WalletAddresses, rpi.Asset)
			if err != nil {
				return nil, err
			}
		}

		// is the farmer selling his own capacity so the pool is either owned by
		// that farmer, or sponsors the pool.
		if farm.ThreebotID == pool.CustomerTid || farm.ThreebotID == pool.SponsorTid {
			distribution = AssetDistributions[DistributionFamerSales]
		}
	}

	var payouts []Payout
	for destination, amount := range distribution {
		if amount == 0 {
			continue
		}
		var address string
		switch destination {
		case FarmerDestination:
			address = farmerAddress
		case SalesDestination:
			address = salesAddress
		case FoundationDestination:
			address = foundationAddress
		case WisdomDestination:
			address = wisdomAddress
		case BurnedDestination:
			address = rpi.Asset.Issuer()
		}

		payout := Payout{
			Address:      address,
			Distribution: amount,
		}
		if err := payout.Valid(); err != nil {
			return nil, errors.Wrapf(err, "payout for '%s' is invlaid", string(destination))
		}

		payouts = append(payouts, payout)
	}

	return payouts, nil
}

// payoutFarmersCap pays out the farmer for a processed reservation
func (e *Stellar) payoutFarmersCap(rpi types.CapacityReservationPaymentInformation) error {
	if rpi.Released || rpi.Canceled || rpi.CancellationPending {
		// already paid
		return nil
	}
	farm, err := e.farmAPI.GetByID(e.ctx, e.db, int64(rpi.FarmerID))
	if err != nil {
		return errors.Wrap(err, "failed to load farm info")
	}

	payouts, err := e.getPayouts(rpi, farm)
	if err != nil {
		return err
	}

	amounts := e.splitPayout(rpi.Amount, payouts)

	paymentInfo := []stellar.PayoutInfo{}
	for i, amount := range amounts {
		paymentInfo = append(
			paymentInfo,
			stellar.PayoutInfo{
				Address: payouts[i].Address,
				Amount:  xdr.Int64(amount),
			},
		)
	}

	addressInfo, err := types.CustomerAddressByAddress(e.ctx, e.db, rpi.Address)
	if err != nil {
		log.Error().Msgf("failed to load escrow address info: %s", err)
		return errors.Wrap(err, "could not load escrow address info")
	}
	if err = e.wallet.QueuePayout(addressInfo.Secret, paymentInfo, capacityReservationMemo(rpi.ReservationID), rpi.Asset, rpi.ReservationID, e.paymentsChannel); err != nil {
		log.Error().Msgf("failed to pay farmer: %s for reservation %d", err, rpi.ReservationID)
		return errors.Wrap(err, "could not pay farmer")
	}
	return nil
}

func (e *Stellar) refundCapacityEscrow(escrowInfo types.CapacityReservationPaymentInformation, cause string) error {
	slog := log.With().
		Str("address", escrowInfo.Address).
		Int64("reservation_id", int64(escrowInfo.ReservationID)).
		Logger()

	slog.Info().Msgf("try to refund client for escrow")

	addressInfo, err := types.CustomerAddressByAddress(e.ctx, e.db, escrowInfo.Address)
	if err != nil {
		return errors.Wrap(err, "failed to load escrow info")
	}
	batchTxs, err := getBatchMemoTransactions(e.ctx, e.db, capacityReservationMemo(escrowInfo.ReservationID))
	if err != nil {
		return errors.Wrap(err, "failed to get memo transactions")
	}
	if err = e.wallet.Refund(addressInfo.Secret, capacityReservationMemo(escrowInfo.ReservationID), escrowInfo.Asset, &batchTxs, e.paymentsChannel, escrowInfo.ReservationID); err != nil {
		return errors.Wrap(err, "failed to refund clients")
	}
	escrowInfo.CancellationPending = true
	escrowInfo.Cause = cause
	if err = types.CapacityReservationPaymentInfoUpdate(e.ctx, e.db, escrowInfo); err != nil {
		return errors.Wrap(err, "failed to mark expired reservation escrow info as cancelled")
	}

	slog.Info().Msgf("refund client scheduled for escrow")
	return nil
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
func (e *Stellar) splitPayout(totalAmount xdr.Int64, payouts []Payout) []int64 {
	// we can't just use big.Float for this calculation, since we need to verify
	// the rounding afterwards

	// sorting is pretty important because the logic below will
	// give the change to the first payout that gets paid
	sort.Slice(payouts, func(i, j int) bool {
		// less
		return payouts[i].Destination < payouts[j].Destination
	})

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
	var change int64
	amounts := make([]int64, 0, len(payouts))
	for _, payout := range payouts {
		amount := baseAmount * int64(payout.Distribution)
		change += amount % multiplier
		amount /= multiplier
		amounts = append(amounts, amount)
	}

	change /= multiplier

	for i := range amounts {
		v := amounts[i]
		if v != 0 {
			amounts[i] += change
			break
		}
	}

	return amounts
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

// getNetworkDivisor gets a divisor for the fee to be paid based on the current
// grid network
func (e *Stellar) getNetworkDivisor() int64 {
	divisor, err := e.gridNetwork.Divisor()
	if err != nil {
		log.Error().Msgf("unknown gridnetwork \"%s\", defaulting to base fee", e.gridNetwork)
		return 1
	}

	return divisor
}

func addressByAsset(addrs []gdirectory.WalletAddress, asset stellar.Asset) (string, error) {
	for _, a := range addrs {
		if a.Asset == asset.Code() && a.Address != "" {
			return a.Address, nil
		}
	}
	return "", fmt.Errorf("not address found for asset %s", asset)
}

func capacityReservationMemo(id schema.ID) string {
	return fmt.Sprintf("p-%d", id)
}
