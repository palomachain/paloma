package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/address"
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrortypes "github.com/cosmos/cosmos-sdk/types/errors"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/x/gravity/types"
)

// Check that our expected keeper types are implemented
var (
	_ types.StakingKeeper      = (*stakingkeeper.Keeper)(nil)
	_ types.SlashingKeeper     = (*slashingkeeper.Keeper)(nil)
	_ types.DistributionKeeper = (*distrkeeper.Keeper)(nil)
)

// Keeper maintains the link to storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	paramSpace paramtypes.Subspace

	cdc               codec.BinaryCodec // The wire codec for binary encoding/decoding.
	bankKeeper        types.BankKeeper
	StakingKeeper     types.StakingKeeper
	SlashingKeeper    types.SlashingKeeper
	DistKeeper        distrkeeper.Keeper
	accountKeeper     types.AccountKeeper
	ibcTransferKeeper ibctransferkeeper.Keeper
	evmKeeper         types.EVMKeeper
	addressCodec      address.Codec
	storeGetter       keeperutil.StoreGetter

	AttestationHandler interface {
		Handle(context.Context, types.Attestation, types.EthereumClaim) error
	}
	authority string
}

// NewKeeper returns a new instance of the gravity keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	accKeeper types.AccountKeeper,
	stakingKeeper types.StakingKeeper,
	bankKeeper types.BankKeeper,
	slashingKeeper types.SlashingKeeper,
	distributionKeeper distrkeeper.Keeper,
	ibcTransferKeeper ibctransferkeeper.Keeper,
	evmKeeper types.EVMKeeper,
	storeGetter keeperutil.StoreGetter,
	authority string,
) Keeper {
	k := Keeper{

		cdc:                cdc,
		bankKeeper:         bankKeeper,
		StakingKeeper:      stakingKeeper,
		SlashingKeeper:     slashingKeeper,
		DistKeeper:         distributionKeeper,
		accountKeeper:      accKeeper,
		ibcTransferKeeper:  ibcTransferKeeper,
		evmKeeper:          evmKeeper,
		storeGetter:        storeGetter,
		AttestationHandler: nil,
	}
	attestationHandler := AttestationHandler{keeper: &k}
	attestationHandler.ValidateMembers()
	k.AttestationHandler = attestationHandler
	k.authority = authority
	return k
}

////////////////////////
/////// HELPERS ////////
////////////////////////

// SendToCommunityPool handles incorrect SendToPaloma calls to the community pool, since the calls
// have already been made on Ethereum there's nothing we can do to reverse them, and we should at least
// make use of the tokens which would otherwise be lost
func (k Keeper) SendToCommunityPool(ctx context.Context, coins sdk.Coins) error {
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, distrtypes.ModuleName, coins); err != nil {
		return sdkerrors.Wrap(err, "transfer to community pool failed")
	}
	feePool, err := k.DistKeeper.FeePool.Get(ctx)
	if err != nil {
		return err
	}
	feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(coins...)...)
	return k.DistKeeper.FeePool.Set(ctx, feePool)
}

/////////////////////////////
//////// PARAMETERS /////////
/////////////////////////////

// GetParamsIfSet returns the parameters from the store if they exist, or an error
// This is useful for certain contexts where the store is not yet set up, like
// in an AnteHandler during InitGenesis
func (k Keeper) GetParamsIfSet(ctx context.Context) (params types.Params, err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, pair := range params.ParamSetPairs() {
		if !k.paramSpace.Has(sdkCtx, pair.Key) {
			return types.Params{}, sdkerrors.Wrapf(sdkerrortypes.ErrNotFound, "the param key %s has not been set", string(pair.Key))
		}
		k.paramSpace.Get(sdkCtx, pair.Key, pair.Value)
	}

	return
}

// GetParams returns the parameters from the store
func (k Keeper) GetParams(ctx context.Context) (params types.Params) {
	bz := k.GetStore(ctx).Get([]byte(types.ParamsKey))
	if bz == nil {
		return params
	}
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the parameters in the store
func (k Keeper) SetParams(ctx context.Context, params types.Params) {
	bz := k.cdc.MustMarshal(&params)
	k.GetStore(ctx).Set([]byte(types.ParamsKey), bz)
}

// GetBridgeContractAddress returns the bridge contract address on ETH
func (k Keeper) GetBridgeContractAddress(ctx context.Context) (*types.EthAddress, error) {
	var a string
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.paramSpace.Get(sdkCtx, types.ParamsStoreKeyBridgeEthereumAddress, &a)
	addr, err := types.NewEthAddress(a)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "found invalid bridge contract address in store: %v", a)
	}
	return addr, nil
}

// GetBridgeChainID returns the chain id of the ETH chain we are running against
func (k Keeper) GetBridgeChainID(ctx context.Context) uint64 {
	var a uint64
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.paramSpace.Get(sdkCtx, types.ParamsStoreKeyBridgeContractChainID, &a)
	return a
}

// Logger returns a module-specific Logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	return liblog.FromSDKLogger(k.Logger(ctx)).With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) UnpackAttestationClaim(att *types.Attestation) (types.EthereumClaim, error) {
	var msg types.EthereumClaim
	err := k.cdc.UnpackAny(att.Claim, &msg)
	if err != nil {
		return nil, err
	} else {
		return msg, nil
	}
}

/////////////////////////////
//////// Parameters /////////
/////////////////////////////

// prefixRange turns a prefix into a (start, end) range. The start is the given prefix value and
// the end is calculated by adding 1 bit to the start value. Nil is not allowed as prefix.
// Example: []byte{1, 3, 4} becomes []byte{1, 3, 5}
// []byte{15, 42, 255, 255} becomes []byte{15, 43, 0, 0}
//
// In case of an overflow the end is set to nil.
// Example: []byte{255, 255, 255, 255} becomes nil
// MARK finish-batches: this is where some crazy shit happens
func prefixRange(prefix []byte) ([]byte, []byte, error) {
	if prefix == nil {
		return nil, nil, fmt.Errorf("nil key not allowed")
	}
	// special case: no prefix is whole range
	if len(prefix) == 0 {
		return nil, nil, nil
	}

	// copy the prefix and update last byte
	end := make([]byte, len(prefix))
	copy(end, prefix)
	l := len(end) - 1
	end[l]++

	// wait, what if that overflowed?....
	for end[l] == 0 && l > 0 {
		l--
		end[l]++
	}

	// okay, funny guy, you gave us FFF, no end to this range...
	if l == 0 && end[0] == 0 {
		end = nil
	}
	return prefix, end, nil
}

// DeserializeValidatorIterator returns validators from the validator iterator.
// Adding here in gravity keeper as cdc is not available inside endblocker.
func (k Keeper) DeserializeValidatorIterator(vals []byte) stakingtypes.ValAddresses {
	validators := stakingtypes.ValAddresses{
		Addresses: []string{},
	}
	k.cdc.MustUnmarshal(vals, &validators)
	return validators
}

// InvalidSendToEthAddress Returns true if the provided address is invalid to send to Ethereum this could be
// for one of several reasons. (1) it is invalid in general like the Zero address, (2)
// it is invalid for a subset of ERC20 addresses. (2) is not yet implemented
// Blocking some addresses is technically motivated, if any ERC20 transfers in a batch fail the entire batch
// becomes impossible to execute.
func (k Keeper) InvalidSendToEthAddress(ctx context.Context, addr types.EthAddress, _erc20Addr types.EthAddress) bool {
	return addr == types.ZeroAddress()
}

func (k Keeper) GetStore(ctx context.Context) storetypes.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return k.storeGetter.Store(sdkCtx)
}

type GravityStoreGetter struct {
	storeKey storetypes.StoreKey
}

func NewGravityStoreGetter(storeKey storetypes.StoreKey) GravityStoreGetter {
	return GravityStoreGetter{
		storeKey: storeKey,
	}
}

func (gsg GravityStoreGetter) Store(ctx context.Context) storetypes.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.KVStore(gsg.storeKey)
}

func (k Keeper) GetAuthority() string {
	return k.authority
}
