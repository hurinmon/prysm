package state

import (
	"fmt"

	"github.com/pkg/errors"
	fieldparams "github.com/prysmaticlabs/prysm/v3/config/fieldparams"
	"github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v3/consensus-types/state/types"
	"github.com/prysmaticlabs/prysm/v3/encoding/bytesutil"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/runtime/version"
)

// ValidatorIndexOutOfRangeError represents an error scenario where a validator does not exist
// at a given index in the validator's array.
type ValidatorIndexOutOfRangeError struct {
	message string
}

// NewValidatorIndexOutOfRangeError creates a new error instance.
func NewValidatorIndexOutOfRangeError(index primitives.ValidatorIndex) ValidatorIndexOutOfRangeError {
	return ValidatorIndexOutOfRangeError{
		message: fmt.Sprintf("index %d out of range", index),
	}
}

// Error returns the underlying error message.
func (e *ValidatorIndexOutOfRangeError) Error() string {
	return e.message
}

// Validators participating in consensus on the beacon chain.
func (b *State) Validators() []*ethpb.Validator {
	if b.validators == nil {
		return nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.validatorsVal()
}

// validatorsVal participating in consensus on the beacon chain.
// This assumes that a lock is already held on BeaconState.
func (b *State) validatorsVal() []*ethpb.Validator {
	if b.validators == nil {
		return nil
	}

	res := make([]*ethpb.Validator, len(b.validators))
	for i := 0; i < len(res); i++ {
		val := b.validators[i]
		if val == nil {
			continue
		}
		res[i] = ethpb.CopyValidator(val)
	}
	return res
}

// references of validators participating in consensus on the beacon chain.
// This assumes that a lock is already held on BeaconState. This does not
// copy fully and instead just copies the reference.
func (b *State) validatorsReferences() []*ethpb.Validator {
	if b.validators == nil {
		return nil
	}

	res := make([]*ethpb.Validator, len(b.validators))
	for i := 0; i < len(res); i++ {
		validator := b.validators[i]
		if validator == nil {
			continue
		}
		// copy validator reference instead.
		res[i] = validator
	}
	return res
}

// ValidatorAtIndex is the validator at the provided index.
func (b *State) ValidatorAtIndex(idx primitives.ValidatorIndex) (*ethpb.Validator, error) {
	if b.validators == nil {
		return &ethpb.Validator{}, nil
	}
	if uint64(len(b.validators)) <= uint64(idx) {
		e := NewValidatorIndexOutOfRangeError(idx)
		return nil, &e
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	val := b.validators[idx]
	return ethpb.CopyValidator(val), nil
}

// ValidatorAtIndexReadOnly is the validator at the provided index. This method
// doesn't clone the validator.
func (b *State) ValidatorAtIndexReadOnly(idx primitives.ValidatorIndex) (types.ReadOnlyValidator, error) {
	if b.validators == nil {
		return nil, ErrNilValidatorsInState
	}
	if uint64(len(b.validators)) <= uint64(idx) {
		e := NewValidatorIndexOutOfRangeError(idx)
		return nil, &e
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return NewValidator(b.validators[idx])
}

// ValidatorIndexByPubkey returns a given validator by its 48-byte public key.
func (b *State) ValidatorIndexByPubkey(key [fieldparams.BLSPubkeyLength]byte) (primitives.ValidatorIndex, bool) {
	if b == nil || b.valMapHandler == nil || b.valMapHandler.IsNil() {
		return 0, false
	}
	b.lock.RLock()
	defer b.lock.RUnlock()
	numOfVals := len(b.validators)

	idx, ok := b.valMapHandler.Get(key)
	if ok && primitives.ValidatorIndex(numOfVals) <= idx {
		return primitives.ValidatorIndex(0), false
	}
	return idx, ok
}

// PubkeyAtIndex returns the pubkey at the given
// validator index.
func (b *State) PubkeyAtIndex(idx primitives.ValidatorIndex) [fieldparams.BLSPubkeyLength]byte {
	if uint64(idx) >= uint64(len(b.validators)) {
		return [fieldparams.BLSPubkeyLength]byte{}
	}
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.validators[idx] == nil {
		return [fieldparams.BLSPubkeyLength]byte{}
	}
	return bytesutil.ToBytes48(b.validators[idx].PublicKey)
}

// NumValidators returns the size of the validator registry.
func (b *State) NumValidators() int {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return len(b.validators)
}

// ReadFromEveryValidator reads values from every validator and applies it to the provided function.
//
// WARNING: This method is potentially unsafe, as it exposes the actual validator registry.
func (b *State) ReadFromEveryValidator(f func(idx int, val types.ReadOnlyValidator) error) error {
	if b.validators == nil {
		return errors.New("nil validators in state")
	}
	b.lock.RLock()
	validators := b.validators
	b.lock.RUnlock()

	for i, v := range validators {
		v, err := NewValidator(v)
		if err != nil {
			return err
		}
		if err := f(i, v); err != nil {
			return err
		}
	}
	return nil
}

// Balances of validators participating in consensus on the beacon chain.
func (b *State) Balances() []uint64 {
	if b.balances == nil {
		return nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.balancesVal()
}

// balancesVal of validators participating in consensus on the beacon chain.
// This assumes that a lock is already held on BeaconState.
func (b *State) balancesVal() []uint64 {
	if b.balances == nil {
		return nil
	}

	res := make([]uint64, len(b.balances))
	copy(res, b.balances)
	return res
}

// BalanceAtIndex of validator with the provided index.
func (b *State) BalanceAtIndex(idx primitives.ValidatorIndex) (uint64, error) {
	if b.balances == nil {
		return 0, nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	if uint64(len(b.balances)) <= uint64(idx) {
		return 0, fmt.Errorf("index of %d does not exist", idx)
	}
	return b.balances[idx], nil
}

// BalancesLength returns the length of the balances slice.
func (b *State) BalancesLength() int {
	if b.balances == nil {
		return 0
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.balancesLength()
}

// Slashings of validators on the beacon chain.
func (b *State) Slashings() []uint64 {
	if b.slashings == nil {
		return nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.slashingsVal()
}

// slashingsVal of validators on the beacon chain.
// This assumes that a lock is already held on BeaconState.
func (b *State) slashingsVal() []uint64 {
	if b.slashings == nil {
		return nil
	}

	res := make([]uint64, len(b.slashings))
	copy(res, b.slashings)
	return res
}

// InactivityScores of validators participating in consensus on the beacon chain.
func (b *State) InactivityScores() ([]uint64, error) {
	if b.version == version.Phase0 {
		return nil, errNotSupported("InactivityScores", b.version)
	}

	if b.inactivityScores == nil {
		return nil, nil
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.inactivityScoresVal(), nil
}

// inactivityScoresVal of validators participating in consensus on the beacon chain.
// This assumes that a lock is already held on BeaconState.
func (b *State) inactivityScoresVal() []uint64 {
	if b.inactivityScores == nil {
		return nil
	}

	res := make([]uint64, len(b.inactivityScores))
	copy(res, b.inactivityScores)
	return res
}