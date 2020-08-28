package workloads

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/threefoldtech/zos/pkg/crypto"

	"github.com/pkg/errors"
)

type (
	Capaciter interface {
		GetRSU() RSU
	}

	Signer interface {
		SignatureChallenge() ([]byte, error)
	}

	Workloader interface {
		GetContract() *Contract
		GetState() *State

		Capaciter
		Signer
	}

	RSU struct {
		CRU int64
		SRU float64
		HRU float64
		MRU float64
	}
)

// SignatureProvisionRequestVerify verify the signature from a signature request
// this is used for provision
// the signature is created from the workload siging challenge + "provision" + customer tid
func SignatureProvisionRequestVerify(w Workloader, pk string, sig SigningSignature) error {
	key, err := crypto.KeyFromHex(pk)
	if err != nil {
		return errors.Wrap(err, "invalid verification key")
	}

	b, err := w.SignatureChallenge()
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	if _, err := buf.WriteString("provision"); err != nil {
		return err
	}
	if _, err := buf.WriteString(fmt.Sprintf("%d", sig.Tid)); err != nil {
		return err
	}

	msg := sha256.Sum256(buf.Bytes())
	signature, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return err
	}

	return crypto.Verify(key, msg[:], signature)
}

// SignatureDeleteRequestVerify verify the signature from a signature request
// this is used for workload delete
// the signature is created from the workload siging challenge + "delete" + customer tid
func SignatureDeleteRequestVerify(w Workloader, pk string, sig SigningSignature) error {
	key, err := crypto.KeyFromHex(pk)
	if err != nil {
		return errors.Wrap(err, "invalid verification key")
	}

	b, err := w.SignatureChallenge()
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	if _, err := buf.WriteString("delete"); err != nil {
		return err
	}
	if _, err := buf.WriteString(fmt.Sprintf("%d", sig.Tid)); err != nil {
		return err
	}

	msg := sha256.Sum256(buf.Bytes())
	signature, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return err
	}

	return crypto.Verify(key, msg[:], signature)
}

// Verify signature
// pk is the public key used as verification key in hex encoded format
// the signature is the signature to verify (in raw binary format)
func Verify(w Workloader, pk string, sig []byte) error {
	key, err := crypto.KeyFromHex(pk)
	if err != nil {
		return errors.Wrap(err, "invalid verification key")
	}

	b, err := w.SignatureChallenge()
	if err != nil {
		return err
	}

	msg := sha256.Sum256(b)

	return crypto.Verify(key, msg[:], sig)
}
