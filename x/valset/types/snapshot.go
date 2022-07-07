package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func (s *Snapshot) GetValidator(find sdk.ValAddress) (Validator, bool) {
	for _, val := range s.Validators {
		if val.Address.Equals(find) {
			return val, true
		}
	}
	return Validator{}, false
}
