package state

import (
	"github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v3/consensus-types/state/types"
	"github.com/prysmaticlabs/prysm/v3/runtime/version"
)

// SetNextWithdrawalIndex sets the index that will be assigned to the next withdrawal.
func (b *State) SetNextWithdrawalIndex(i uint64) error {
	if b.version < version.Capella {
		return errNotSupported("SetNextWithdrawalIndex", b.version)
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	b.nextWithdrawalIndex = i
	b.markFieldAsDirty(types.NextWithdrawalIndex)
	return nil
}

// SetNextWithdrawalValidatorIndex sets the index of the validator which is
// next in line for a partial withdrawal.
func (b *State) SetNextWithdrawalValidatorIndex(i primitives.ValidatorIndex) error {
	if b.version < version.Capella {
		return errNotSupported("SetNextWithdrawalValidatorIndex", b.version)
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	b.nextWithdrawalValidatorIndex = i
	b.markFieldAsDirty(types.NextWithdrawalValidatorIndex)
	return nil
}