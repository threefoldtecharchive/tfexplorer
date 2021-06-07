package stellar

import (
	"fmt"
	"math/big"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/rs/zerolog/log"
	"github.com/stellar/go/amount"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	hProtocol "github.com/stellar/go/protocols/horizon"
	horizoneffects "github.com/stellar/go/protocols/horizon/effects"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/go/xdr"
)

type (
	// Signers is a flag type for setting the signers on the escrow accounts
	Signers []string

	// PayoutInfo holds information about which address needs to receive how many funds
	// for payment commands which take multiple receivers
	PayoutInfo struct {
		Address string
		Amount  xdr.Int64
	}

	// stellarWallet is the foundation wallet
	// Payments will be funded and fees will be taken with this wallet
	stellarWallet struct {
		keypair    *keypair.Full
		network    string
		assets     map[Asset]struct{}
		signers    Signers
		horizonURL string
	}

	// Wallet interface
	Wallet interface {
		AssetFromCode(code string) (Asset, error)
		PrecisionDigits() int
		PublicAddress() string
		CreateAccount() (encSeed string, address string, err error)
		GetBalance(address string, memo string, asset Asset) (xdr.Int64, []string, error)
		Refund(encryptedSeed string, memo string, asset Asset) error
		PayoutFarmers(encryptedSeed string, destinations []PayoutInfo, memo string, asset Asset) error
		GetAccountDetails(address string) (account hProtocol.Account, err error)
		GetHorizonClient() (*horizonclient.Client, error)
		GetNetworkPassPhrase() string
	}
)

const (
	stellarPrecision       = 1e7
	stellarPrecisionDigits = 7
	stellarPageLimit       = 200
	stellarOneCoin         = 10000000

	// NetworkProduction uses stellar production network
	NetworkProduction = "production"
	// NetworkTest uses stellar test network
	NetworkTest = "testnet"
	// NetworkDebug doesn't do validation, and always address validation is skipped
	// Only supported by the AddressValidator
	NetworkDebug = "debug"

	// global multiplier for the base fee
	baseFeeMultiplier = 2
)

var (
	// ErrInsufficientBalance is an error that is used when there is insufficient balance
	ErrInsufficientBalance = errors.New("insufficient balance")
	// ErrAssetCodeNotSupported indicated the given asset code is not supported by this wallet
	ErrAssetCodeNotSupported = errors.New("asset code not supported")
)

// New stellar wallet from an optional seed. If no seed is given (i.e. empty string),
// the wallet will panic on all actions which need to be signed, or otherwise require
// a key to be loaded.
func New(seed, network string, signers []string, horizonURL string) (Wallet, error) {
	assets := mainnetAssets

	if len(signers) < 3 && seed != "" {
		log.Warn().Msg("to enable escrow account recovery, provide at least 3 signers")
	}

	w := &stellarWallet{
		network:    network,
		assets:     assets,
		signers:    signers,
		horizonURL: horizonURL,
	}

	var err error
	if seed != "" {
		w.keypair, err = keypair.ParseFull(seed)
		if err != nil {
			return nil, err
		}
	}

	return &retryWallet{w}, nil
}

// AssetFromCode loads the full asset from a code, provided the wallet supports
// the asset code
func (w *stellarWallet) AssetFromCode(code string) (Asset, error) {
	for asset := range w.assets {
		if asset.Code() == code {
			return asset, nil
		}
	}
	return "", ErrAssetCodeNotSupported
}

// PrecisionDigits of the underlying currencies on chain
func (w *stellarWallet) PrecisionDigits() int {
	return stellarPrecisionDigits
}

// PublicAddress of this wallet
func (w *stellarWallet) PublicAddress() string {
	if w.keypair == nil {
		return ""
	}
	return w.keypair.Address()
}

// CreateAccount and activate it, so that it is ready to be used
// The encrypted seed of the wallet is returned, together with the public address
func (w *stellarWallet) CreateAccount() (string, string, error) {
	client, err := w.GetHorizonClient()
	if err != nil {
		return "", "", err
	}
	newKp, err := keypair.Random()
	if err != nil {
		return "", "", err
	}

	feeFactor := int64(1)

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = time.Minute // retry for 1 min maximum
	bo.MaxInterval = time.Second * 1

	err = backoff.RetryNotify(func() error {
		sourceAccount, err := w.GetAccountDetails(w.keypair.Address())
		if err != nil {
			return backoff.Permanent(errors.Wrap(err, "failed to get source account"))
		}

		err = w.activateEscrowAccount(newKp, sourceAccount, client, feeFactor)
		if err != nil {
			return errors.Wrapf(err, "failed to activate escrow account %s", newKp.Address())
		}

		return nil
	}, bo, func(err error, d time.Duration) {
		if feeFactor < 5 {
			feeFactor++
		}
		log.Error().
			Err(err).
			Str("sleep", d.String()).
			Msgf("failed submit activate escrow account transaction for: %s", newKp.Address())
	})

	if err != nil {
		return "", "", err
	}

	log.Info().Msg("escrow account activation successful")

	bo = backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = time.Minute // retry for 1 min maximum
	bo.MaxInterval = time.Second * 1

	err = backoff.Retry(func() error {
		// Now fetch the escrow source account to perform operations on it
		sourceAccount, err := w.GetAccountDetails(newKp.Address())
		if err != nil {
			return errors.Wrap(err, "failed to get escrow source account")
		}

		err = w.setupEscrow(newKp, sourceAccount, client)
		if err != nil {
			return errors.Wrapf(err, "failed to setup escrow account %s", newKp.Address())
		}

		return nil
	}, bo)

	if err != nil {
		return "", "", err
	}

	// encrypt the seed before it is returned
	encryptedSeed, err := encrypt(newKp.Seed(), w.encryptionKey())
	if err != nil {
		return "", "", errors.Wrap(err, "could not encrypt new wallet seed")
	}

	return encryptedSeed, newKp.Address(), nil
}

func (w *stellarWallet) activateEscrowAccount(newKp *keypair.Full, sourceAccount hProtocol.Account, client *horizonclient.Client, feeFactor int64) error {
	currency := big.NewRat(int64(w.getMinumumBalance()), stellarPrecision)
	minimumBalance := currency.FloatString(stellarPrecisionDigits)
	createAccountOp := txnbuild.CreateAccount{
		Destination: newKp.Address(),
		Amount:      minimumBalance,
	}
	ops := []txnbuild.Operation{&createAccountOp}

	tx, err := txnbuild.NewTransaction(
		txnbuild.TransactionParams{
			SourceAccount:        &sourceAccount,
			Operations:           ops,
			Timebounds:           txnbuild.NewTimeout(300),
			IncrementSequenceNum: true,
			BaseFee:              (txnbuild.MinBaseFee * int64(len(ops))) * int64(feeFactor) * baseFeeMultiplier,
		},
	)
	if err != nil {
		return backoff.Permanent(errors.Wrap(err, "failed to build transaction"))
	}

	tx, err = tx.Sign(w.GetNetworkPassPhrase(), w.keypair)
	if err != nil {
		return backoff.Permanent(errors.Wrap(err, "failed to sign transaction"))
	}

	log.Info().Msg("trying to submit activate escrow account transaction")
	_, err = client.SubmitTransaction(tx)
	if err != nil {
		if hError, ok := err.(*horizonclient.Error); ok {
			log.Error().
				Err(err).
				Str("problem", fmt.Sprintf("%+v", hError.Problem.Extras)).
				Msg("error submitting transaction for account activation")
		}

		return err
	}

	return nil
}

// setupEscrow will setup a trustline to the correct asset and issuer
// and also setup multisig on the escrow
func (w *stellarWallet) setupEscrow(newKp *keypair.Full, sourceAccount hProtocol.Account, client *horizonclient.Client) error {
	var operations []txnbuild.Operation

	trustlineOperation := w.setupTrustline(sourceAccount)
	operations = append(operations, trustlineOperation...)

	addSignerOperations := w.setupEscrowMultisig(sourceAccount)
	if addSignerOperations != nil {
		operations = append(operations, addSignerOperations...)
	}

	tx := txnbuild.TransactionParams{
		Operations: operations,
		Timebounds: txnbuild.NewTimeout(300),
	}

	fundedTx, err := w.fundTransaction(&tx)
	if err != nil {
		return errors.Wrap(err, "failed to fund transaction")
	}

	err = w.signAndSubmitTx(newKp, fundedTx)
	if err != nil {
		return errors.Wrap(err, "failed to sign and submit transaction")
	}
	return nil
}

func (w *stellarWallet) setupTrustline(sourceAccount hProtocol.Account) []txnbuild.Operation {
	ops := make([]txnbuild.Operation, 0, len(w.assets))
	for asset := range w.assets {
		ops = append(ops, &txnbuild.ChangeTrust{
			SourceAccount: &sourceAccount,
			Line: txnbuild.CreditAsset{
				Code:   asset.Code(),
				Issuer: asset.Issuer(),
			},
		})
	}
	return ops
}

func (w *stellarWallet) setupEscrowMultisig(sourceAccount hProtocol.Account) []txnbuild.Operation {
	if len(w.signers) < 3 {
		// not enough signers, don't add multisig
		return nil
	}
	// set the threshold for the master key equal to the amount of signers
	threshold := txnbuild.Threshold(len(w.signers))

	// set the threshold to complete transaction for signers. atleast 3 signatures are required
	txThreshold := txnbuild.Threshold(3)
	if len(w.signers) > 3 {
		txThreshold = txnbuild.Threshold(len(w.signers)/2 + 1)
	}

	var operations []txnbuild.Operation
	// add the signing options
	addSignersOp := txnbuild.SetOptions{
		SourceAccount:   &sourceAccount,
		LowThreshold:    txnbuild.NewThreshold(0),
		MediumThreshold: txnbuild.NewThreshold(txThreshold),
		HighThreshold:   txnbuild.NewThreshold(txThreshold),
		MasterWeight:    txnbuild.NewThreshold(threshold),
	}
	operations = append(operations, &addSignersOp)

	// add the signers
	for _, signer := range w.signers {
		addSignerOperation := txnbuild.SetOptions{
			SourceAccount: &sourceAccount,
			Signer: &txnbuild.Signer{
				Address: signer,
				Weight:  1,
			},
		}
		operations = append(operations, &addSignerOperation)
	}

	return operations
}

// GetBalance gets balance for an address and a given reservation id. It also returns
// a list of addresses which funded the given address.
func (w *stellarWallet) GetBalance(address string, memo string, asset Asset) (xdr.Int64, []string, error) {
	if address == "" {
		err := fmt.Errorf("trying to get the balance of an empty address. this should never happen")
		log.Warn().Err(err).Send()
		return 0, nil, err
	}

	var total xdr.Int64
	horizonClient, err := w.GetHorizonClient()
	if err != nil {
		return 0, nil, err
	}

	cursor := ""

	txReq := horizonclient.TransactionRequest{
		ForAccount: address,
		Cursor:     cursor,
		Limit:      stellarPageLimit,
	}

	log.Info().Str("address", address).Msg("fetching balance for address")
	txes, err := horizonClient.Transactions(txReq)
	if err != nil {
		return 0, nil, errors.Wrap(err, "could not get transactions")
	}

	donors := make(map[string]struct{})
	for len(txes.Embedded.Records) != 0 {
		for _, tx := range txes.Embedded.Records {
			if tx.Memo == memo {
				effectsReq := horizonclient.EffectRequest{
					ForTransaction: tx.Hash,
				}
				effects, err := horizonClient.Effects(effectsReq)
				if err != nil {
					log.Error().Err(err).Msgf("failed to get transaction effects")
					continue
				}
				// first check if we have been paid
				var isFunding bool
				for _, effect := range effects.Embedded.Records {
					if effect.GetAccount() != address {
						continue
					}
					if effect.GetType() == "account_credited" {
						creditedEffect := effect.(horizoneffects.AccountCredited)
						if creditedEffect.Asset.Code != asset.Code() ||
							creditedEffect.Asset.Issuer != asset.Issuer() {
							continue
						}
						parsedAmount, err := amount.Parse(creditedEffect.Amount)
						if err != nil {
							continue
						}
						isFunding = true
						total += parsedAmount
					} else if effect.GetType() == "account_debited" {
						debitedEffect := effect.(horizoneffects.AccountDebited)
						if debitedEffect.Asset.Code != asset.Code() ||
							debitedEffect.Asset.Issuer != asset.Issuer() {
							continue
						}
						parsedAmount, err := amount.Parse(debitedEffect.Amount)
						if err != nil {
							continue
						}
						isFunding = false
						total -= parsedAmount
					}
				}
				if isFunding {
					// we don't need to verify the asset here anymore, since this
					// flag is only toggled on after that check passed in the loop
					// above
					for _, effect := range effects.Embedded.Records {
						if effect.GetType() == "account_debited" && effect.GetAccount() != address {
							donors[effect.GetAccount()] = struct{}{}
						}
					}
				}
			}
			cursor = tx.PagingToken()
		}

		// if the amount of records fetched is smaller than the page limit
		// we can assume we are on the last page and we break to prevent another
		// call to horizon
		if len(txes.Embedded.Records) < stellarPageLimit {
			break
		}

		txReq.Cursor = cursor
		log.Info().Str("address", address).Msgf("fetching balance for address with cursor: %s", cursor)
		txes, err = horizonClient.Transactions(txReq)
		if err != nil {
			return 0, nil, errors.Wrap(err, "could not get transactions")
		}
	}

	donorList := []string{}
	for donor := range donors {
		donorList = append(donorList, donor)
	}
	log.Info().
		Int64("balance", int64(total)).
		Str("address", address).
		Str("memo", memo).Msgf("status of balance for reservation")
	return total, donorList, nil
}

// Refund an escrow address for a reservation. This will transfer all funds
// for this reservation that are currently on the address (if any), to (some of)
// the addresses which these funds came from.
func (w *stellarWallet) Refund(encryptedSeed string, memo string, asset Asset) error {
	keypair, err := w.keypairFromEncryptedSeed(encryptedSeed)
	if err != nil {
		return errors.Wrap(err, "could not get keypair from encrypted seed")
	}

	amount, funders, err := w.GetBalance(keypair.Address(), memo, asset)
	if err != nil {
		return errors.Wrap(err, "failed to get balance")
	}
	// if no balance for this reservation, do nothing
	if amount == 0 {
		return nil
	}

	sourceAccount, err := w.GetAccountDetails(keypair.Address())
	if err != nil {
		return errors.Wrap(err, "failed to get source account")
	}

	destination := funders[0]

	paymentOP := txnbuild.Payment{
		Destination: destination,
		Amount:      big.NewRat(int64(amount), stellarPrecision).FloatString(stellarPrecisionDigits),
		Asset: txnbuild.CreditAsset{
			Code:   asset.Code(),
			Issuer: asset.Issuer(),
		},
		SourceAccount: &sourceAccount,
	}

	memoText := txnbuild.MemoText(memo)
	tx := txnbuild.TransactionParams{
		Operations: []txnbuild.Operation{&paymentOP},
		Timebounds: txnbuild.NewTimeout(300),
		Memo:       memoText,
	}

	fundedTx, err := w.fundTransaction(&tx)
	if err != nil {
		return errors.Wrap(err, "failed to fund transaction")
	}

	log.Debug().Int64("amount", int64(amount)).Str("destination", destination).Msg("refund")
	err = w.signAndSubmitTx(&keypair, fundedTx)
	if err != nil {
		return errors.Wrap(err, "failed to sign and submit transaction")
	}
	return nil
}

// PayoutFarmers pays a group of farmers, from an escrow account. The escrow
// account must be provided as the encrypted string of the seed.
func (w *stellarWallet) PayoutFarmers(encryptedSeed string, destinations []PayoutInfo, memo string, asset Asset) error {
	keypair, err := w.keypairFromEncryptedSeed(encryptedSeed)
	if err != nil {
		return errors.Wrap(err, "could not get keypair from encrypted seed")
	}
	sourceAccount, err := w.GetAccountDetails(keypair.Address())
	if err != nil {
		return errors.Wrap(err, "failed to get source account")
	}

	paymentOps := make([]txnbuild.Operation, 0, len(destinations)+1)

	for _, pi := range destinations {
		paymentOps = append(paymentOps, &txnbuild.Payment{
			Destination: pi.Address,
			Amount:      big.NewRat(int64(pi.Amount), stellarPrecision).FloatString(stellarPrecisionDigits),
			Asset: txnbuild.CreditAsset{
				Code:   asset.Code(),
				Issuer: asset.Issuer(),
			},
			SourceAccount: &sourceAccount,
		})
	}

	memoText := txnbuild.MemoText(memo)
	tx := txnbuild.TransactionParams{
		Operations: paymentOps,
		Timebounds: txnbuild.NewTimeout(300),
		Memo:       memoText,
	}

	fundedTx, err := w.fundTransaction(&tx)
	if err != nil {
		return errors.Wrap(err, "failed to fund transaction")
	}

	err = w.signAndSubmitTx(&keypair, fundedTx)
	if err != nil {
		return errors.Wrap(err, "failed to sign and submit transaction")
	}
	return nil
}

// fundTransaction funds a transaction with the foundation wallet
// For every operation in the transaction, the fee will be paid by the foundation wallet
func (w *stellarWallet) fundTransaction(txp *txnbuild.TransactionParams) (*txnbuild.Transaction, error) {
	sourceAccount, err := w.GetAccountDetails(w.keypair.Address())
	if err != nil {
		return &txnbuild.Transaction{}, errors.Wrap(err, "failed to get source account")
	}

	// set the source account of the tx to the foundation account
	txp.SourceAccount = &sourceAccount

	if len(txp.Operations) == 0 {
		return &txnbuild.Transaction{}, errors.New("no operations were set on the transaction")
	}

	// calculate total fee based on the operations in the transaction
	txp.BaseFee = txnbuild.MinBaseFee * int64(len(txp.Operations)) * baseFeeMultiplier
	txp.IncrementSequenceNum = true

	tx, err := txnbuild.NewTransaction(*txp)
	if err != nil {
		return &txnbuild.Transaction{}, errors.Wrap(err, "failed to build transaction")
	}

	tx, err = tx.Sign(w.GetNetworkPassPhrase(), w.keypair)
	if err != nil {
		return &txnbuild.Transaction{}, errors.Wrap(err, "failed to sign transaction")
	}

	return tx, nil
}

// signAndSubmitTx sings of on a transaction with a given keypair
// and submits it to the network
func (w *stellarWallet) signAndSubmitTx(keypair *keypair.Full, tx *txnbuild.Transaction) error {
	client, err := w.GetHorizonClient()
	if err != nil {
		return errors.Wrap(err, "failed to get horizon client")
	}

	tx, err = tx.Sign(w.GetNetworkPassPhrase(), keypair)
	if err != nil {
		return errors.Wrap(err, "failed to sign transaction with keypair")
	}

	log.Info().Msg("submitting transaction to the stellar network")
	// Submit the transaction
	_, err = client.SubmitTransaction(tx)
	if err != nil {
		return errors.Wrap(err, "error submitting transaction")
	}
	return nil
}

// GetAccountDetails gets account details based an a Stellar address
func (w *stellarWallet) GetAccountDetails(address string) (account hProtocol.Account, err error) {
	client, err := w.GetHorizonClient()
	if err != nil {
		return hProtocol.Account{}, err
	}
	ar := horizonclient.AccountRequest{AccountID: address}
	log.Info().Str("address", address).Msgf("fetching account details for address: ")
	account, err = client.AccountDetail(ar)
	if err != nil {
		return hProtocol.Account{}, errors.Wrapf(err, "failed to get account details for account: %s", address)
	}
	return account, nil
}

func (w *stellarWallet) keypairFromEncryptedSeed(seed string) (keypair.Full, error) {
	plainSeed, err := decrypt(seed, w.encryptionKey())
	if err != nil {
		return keypair.Full{}, errors.Wrap(err, "could not decrypt seed")
	}

	kp, err := keypair.ParseFull(plainSeed)
	if err != nil {
		return keypair.Full{}, errors.Wrap(err, "could not parse seed")
	}

	return *kp, nil
}

// GetHorizonClient gets the horizon client based on the wallet's network
func (w *stellarWallet) GetHorizonClient() (*horizonclient.Client, error) {
	if w.horizonURL != "" {
		return &horizonclient.Client{HorizonURL: w.horizonURL}, nil
	}

	switch w.network {
	case "testnet":
		return horizonclient.DefaultTestNetClient, nil
	case "production":
		return horizonclient.DefaultPublicNetClient, nil
	default:
		return nil, errors.New("network is not supported")
	}
}

// GetNetworkPassPhrase gets the Stellar network passphrase based on the wallet's network
func (w *stellarWallet) GetNetworkPassPhrase() string {
	switch w.network {
	case "testnet":
		return network.TestNetworkPassphrase
	case "production":
		return network.PublicNetworkPassphrase
	default:
		return network.TestNetworkPassphrase
	}
}

// getMinumumBalance calculates minimum balance for an escrow account
// following formula is used: minimum Balance = (2 + # of entries) * base reserve
// entries is the amount of operations are required to setup the account
func (w *stellarWallet) getMinumumBalance() int {
	return (2 + len(w.assets) + len(w.signers)) * (stellarOneCoin / 2)
}

func (i *Signers) String() string {
	repr := ""
	for _, s := range *i {
		repr += fmt.Sprintf("%s ", s)
	}
	return repr
}

// Set a value on the signers flag
func (i *Signers) Set(value string) error {
	*i = append(*i, value)
	return nil
}
