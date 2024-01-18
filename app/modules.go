package app

import (
	"encoding/json"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BankModule defines a custom wrapper around the x/bank module's AppModuleBasic
// implementation to provide custom default genesis state.
type BankModule struct {
	bank.AppModuleBasic
}

// DefaultGenesis returns custom Paloma x/bank module genesis state.
func (BankModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	palomaMetadata := banktypes.Metadata{
		Description: "The native staking token of the Paloma network.",
		Base:        BondDenom,
		Name:        DisplayDenom,
		Display:     DisplayDenom,
		Symbol:      strings.ToUpper(DisplayDenom),
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    BondDenom,
				Exponent: 0,
				Aliases: []string{
					"micrograin",
				},
			},
			{
				Denom:    MilliDenom,
				Exponent: 3,
				Aliases: []string{
					"milligrain",
				},
			},
			{
				Denom:    DisplayDenom,
				Exponent: 6,
				Aliases:  []string{},
			},
		},
	}

	genState := banktypes.DefaultGenesisState()
	genState.DenomMetadata = append(genState.DenomMetadata, palomaMetadata)

	return cdc.MustMarshalJSON(genState)
}

// StakingModule defines a custom wrapper around the x/staking module's
// AppModuleBasic implementation to provide custom default genesis state.
type StakingModule struct {
	staking.AppModuleBasic
}

// DefaultGenesis returns custom Paloma x/staking module genesis state.
func (StakingModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	params := stakingtypes.DefaultParams()
	params.BondDenom = BondDenom
	params.MaxValidators = 175

	return cdc.MustMarshalJSON(&stakingtypes.GenesisState{
		Params: params,
	})
}

// CrisisModule defines a custom wrapper around the x/crisis module's
// AppModuleBasic implementation to provide custom default genesis state.
type CrisisModule struct {
	crisis.AppModuleBasic
}

// DefaultGenesis returns custom Paloma x/crisis module genesis state.
func (CrisisModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&crisistypes.GenesisState{
		ConstantFee: sdk.NewCoin(BondDenom, sdkmath.NewInt(1000)),
	})
}

// MintModule defines a custom wrapper around the x/mint module's
// AppModuleBasic implementation to provide custom default genesis state.
type MintModule struct {
	mint.AppModuleBasic
}

// DefaultGenesis returns custom Paloma x/mint module genesis state.
func (MintModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genState := minttypes.DefaultGenesisState()
	genState.Params.MintDenom = BondDenom

	return cdc.MustMarshalJSON(genState)
}

// GovModule defines a custom wrapper around the x/gov module's
// AppModuleBasic implementation to provide custom default genesis state.
type GovModule struct {
	gov.AppModuleBasic
}

// DefaultGenesis returns custom Paloma x/gov module genesis state.
func (GovModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	minDeposit := sdk.NewCoins(sdk.NewCoin(BondDenom, govv1types.DefaultMinDepositTokens))
	genState := govv1types.DefaultGenesisState()
	genState.Params.MinDeposit = minDeposit

	return cdc.MustMarshalJSON(genState)
}
