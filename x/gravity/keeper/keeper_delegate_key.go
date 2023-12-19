package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/gravity/types"
)

///////////////////////////
//// ADDRESS DELEGATION ///
///////////////////////////

// GetOrchestratorValidator returns the validator key associated with an orchestrator key
func (k Keeper) GetOrchestratorValidator(ctx context.Context, orch sdk.AccAddress) (validator stakingtypes.Validator, found bool, err error) {
	valAddr := sdk.ValAddress(orch)
	if valAddr == nil {
		return validator, false, fmt.Errorf("nil valAddr")
	}

	if err := sdk.VerifyAddressFormat(orch); err != nil {
		liblog.FromSDKLogger(k.Logger(ctx)).WithError(err).Error("invalid orch address")
		return validator, false, err
	}

	validator, err = k.StakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		liblog.FromSDKLogger(k.Logger(ctx)).WithError(err).Error("unable to get validator")
		return validator, false, err
	}

	return validator, true, nil
}

/////////////////////////////
// ETH ADDRESS       //
/////////////////////////////

// GetEthAddressByValidator returns the eth address for a given gravity validator
func (k Keeper) GetEthAddressByValidator(ctx context.Context, validator sdk.ValAddress, chainReferenceId string) (ethAddress *types.EthAddress, found bool, err error) {
	return k.evmKeeper.GetEthAddressByValidator(ctx, validator, chainReferenceId)
}

// GetValidatorByEthAddress returns the validator for a given eth address
func (k Keeper) GetValidatorByEthAddress(ctx context.Context, ethAddr types.EthAddress, chainReferenceId string) (validator stakingtypes.Validator, found bool, err error) {
	valAddr, found, err := k.evmKeeper.GetValidatorAddressByEthAddress(ctx, ethAddr, chainReferenceId)
	if err != nil {
		return validator, false, err
	}
	if valAddr == nil {
		return validator, false, nil
	}
	validator, err = k.StakingKeeper.GetValidator(ctx, valAddr)
	return validator, found, nil
}
