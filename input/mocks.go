package input

import (
	"crypto/sha256"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcec/v2/schnorr/musig2"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/lightningnetwork/lnd/keychain"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/stretchr/testify/mock"
)

// MockInput implements the `Input` interface and is used by other packages for
// mock testing.
type MockInput struct {
	mock.Mock
}

// Compile time assertion that MockInput implements Input.
var _ Input = (*MockInput)(nil)

// Outpoint returns the reference to the output being spent, used to construct
// the corresponding transaction input.
func (m *MockInput) OutPoint() wire.OutPoint {
	args := m.Called()
	op := args.Get(0)

	return op.(wire.OutPoint)
}

// RequiredTxOut returns a non-nil TxOut if input commits to a certain
// transaction output. This is used in the SINGLE|ANYONECANPAY case to make
// sure any presigned input is still valid by including the output.
func (m *MockInput) RequiredTxOut() *wire.TxOut {
	args := m.Called()
	txOut := args.Get(0)

	if txOut == nil {
		return nil
	}

	return txOut.(*wire.TxOut)
}

// RequiredLockTime returns whether this input commits to a tx locktime that
// must be used in the transaction including it.
func (m *MockInput) RequiredLockTime() (uint32, bool) {
	args := m.Called()

	return args.Get(0).(uint32), args.Bool(1)
}

// WitnessType returns an enum specifying the type of witness that must be
// generated in order to spend this output.
func (m *MockInput) WitnessType() WitnessType {
	args := m.Called()

	wt := args.Get(0)
	if wt == nil {
		return nil
	}

	return wt.(WitnessType)
}

// SignDesc returns a reference to a spendable output's sign descriptor, which
// is used during signing to compute a valid witness that spends this output.
func (m *MockInput) SignDesc() *SignDescriptor {
	args := m.Called()

	sd := args.Get(0)
	if sd == nil {
		return nil
	}

	return sd.(*SignDescriptor)
}

// CraftInputScript returns a valid set of input scripts allowing this output
// to be spent. The returns input scripts should target the input at location
// txIndex within the passed transaction. The input scripts generated by this
// method support spending p2wkh, p2wsh, and also nested p2sh outputs.
func (m *MockInput) CraftInputScript(signer Signer, txn *wire.MsgTx,
	hashCache *txscript.TxSigHashes,
	prevOutputFetcher txscript.PrevOutputFetcher,
	txinIdx int) (*Script, error) {

	args := m.Called(signer, txn, hashCache, prevOutputFetcher, txinIdx)

	s := args.Get(0)
	if s == nil {
		return nil, args.Error(1)
	}

	return s.(*Script), args.Error(1)
}

// BlocksToMaturity returns the relative timelock, as a number of blocks, that
// must be built on top of the confirmation height before the output can be
// spent. For non-CSV locked inputs this is always zero.
func (m *MockInput) BlocksToMaturity() uint32 {
	args := m.Called()

	return args.Get(0).(uint32)
}

// HeightHint returns the minimum height at which a confirmed spending tx can
// occur.
func (m *MockInput) HeightHint() uint32 {
	args := m.Called()

	return args.Get(0).(uint32)
}

// UnconfParent returns information about a possibly unconfirmed parent tx.
func (m *MockInput) UnconfParent() *TxInfo {
	args := m.Called()

	info := args.Get(0)
	if info == nil {
		return nil
	}

	return info.(*TxInfo)
}

// MockWitnessType implements the `WitnessType` interface and is used by other
// packages for mock testing.
type MockWitnessType struct {
	mock.Mock
}

// Compile time assertion that MockWitnessType implements WitnessType.
var _ WitnessType = (*MockWitnessType)(nil)

// String returns a human readable version of the WitnessType.
func (m *MockWitnessType) String() string {
	args := m.Called()

	return args.String(0)
}

// WitnessGenerator will return a WitnessGenerator function that an output uses
// to generate the witness and optionally the sigScript for a sweep
// transaction.
func (m *MockWitnessType) WitnessGenerator(signer Signer,
	descriptor *SignDescriptor) WitnessGenerator {

	args := m.Called()

	return args.Get(0).(WitnessGenerator)
}

// SizeUpperBound returns the maximum length of the witness of this WitnessType
// if it would be included in a tx. It also returns if the output itself is a
// nested p2sh output, if so then we need to take into account the extra
// sigScript data size.
func (m *MockWitnessType) SizeUpperBound() (lntypes.WeightUnit, bool, error) {
	args := m.Called()

	return args.Get(0).(lntypes.WeightUnit), args.Bool(1), args.Error(2)
}

// AddWeightEstimation adds the estimated size of the witness in bytes to the
// given weight estimator.
func (m *MockWitnessType) AddWeightEstimation(e *TxWeightEstimator) error {
	args := m.Called()

	return args.Error(0)
}

// MockInputSigner is a mock implementation of the Signer interface.
type MockInputSigner struct {
	mock.Mock
}

// Compile-time constraint to ensure MockInputSigner implements Signer.
var _ Signer = (*MockInputSigner)(nil)

// SignOutputRaw generates a signature for the passed transaction according to
// the data within the passed SignDescriptor.
func (m *MockInputSigner) SignOutputRaw(tx *wire.MsgTx,
	signDesc *SignDescriptor) (Signature, error) {

	args := m.Called(tx, signDesc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(Signature), args.Error(1)
}

// ComputeInputScript generates a complete InputIndex for the passed
// transaction with the signature as defined within the passed SignDescriptor.
func (m *MockInputSigner) ComputeInputScript(tx *wire.MsgTx,
	signDesc *SignDescriptor) (*Script, error) {

	args := m.Called(tx, signDesc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*Script), args.Error(1)
}

// MuSig2CreateSession creates a new MuSig2 signing session using the local key
// identified by the key locator.
func (m *MockInputSigner) MuSig2CreateSession(version MuSig2Version,
	locator keychain.KeyLocator, pubkey []*btcec.PublicKey,
	tweak *MuSig2Tweaks, pubNonces [][musig2.PubNonceSize]byte,
	nonces *musig2.Nonces) (*MuSig2SessionInfo, error) {

	args := m.Called(version, locator, pubkey, tweak, pubNonces, nonces)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*MuSig2SessionInfo), args.Error(1)
}

// MuSig2RegisterNonces registers one or more public nonces of other signing
// participants for a session identified by its ID.
func (m *MockInputSigner) MuSig2RegisterNonces(versio MuSig2SessionID,
	pubNonces [][musig2.PubNonceSize]byte) (bool, error) {

	args := m.Called(versio, pubNonces)
	if args.Get(0) == nil {
		return false, args.Error(1)
	}

	return args.Bool(0), args.Error(1)
}

// MuSig2Sign creates a partial signature using the local signing key that was
// specified when the session was created.
func (m *MockInputSigner) MuSig2Sign(sessionID MuSig2SessionID,
	msg [sha256.Size]byte, withSortedKeys bool) (
	*musig2.PartialSignature, error) {

	args := m.Called(sessionID, msg, withSortedKeys)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*musig2.PartialSignature), args.Error(1)
}

// MuSig2CombineSig combines the given partial signature(s) with the local one,
// if it already exists.
func (m *MockInputSigner) MuSig2CombineSig(sessionID MuSig2SessionID,
	partialSig []*musig2.PartialSignature) (
	*schnorr.Signature, bool, error) {

	args := m.Called(sessionID, partialSig)
	if args.Get(0) == nil {
		return nil, false, args.Error(2)
	}

	return args.Get(0).(*schnorr.Signature), args.Bool(1), args.Error(2)
}

// MuSig2Cleanup removes a session from memory to free up resources.
func (m *MockInputSigner) MuSig2Cleanup(sessionID MuSig2SessionID) error {
	args := m.Called(sessionID)

	return args.Error(0)
}
