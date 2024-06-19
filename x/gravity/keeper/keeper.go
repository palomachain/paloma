package keeper

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"cosmossdk.io/core/address"
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
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
	cdc               codec.BinaryCodec // The wire codec for binary encoding/decoding.
	bankKeeper        types.BankKeeper
	StakingKeeper     types.StakingKeeper
	SlashingKeeper    types.SlashingKeeper
	DistKeeper        distrkeeper.Keeper
	accountKeeper     types.AccountKeeper
	ibcTransferKeeper ibctransferkeeper.Keeper
	evmKeeper         types.EVMKeeper
	consensusKeeper   types.ConsensusKeeper
	AddressCodec      address.Codec
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
	consensusKeeper types.ConsensusKeeper,
	storeGetter keeperutil.StoreGetter,
	authority string,
	valAddressCodec address.Codec,
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
		consensusKeeper:    consensusKeeper,
		storeGetter:        storeGetter,
		AttestationHandler: nil,
		AddressCodec:       valAddressCodec,
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

// GetParams returns the parameters from the store
func (k Keeper) GetParams(ctx context.Context) (params types.Params) {
	bz := k.GetStore(ctx, types.StoreModulePrefix).Get([]byte(types.ParamsKey))
	if bz == nil {
		return params
	}
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the parameters in the store
func (k Keeper) SetParams(ctx context.Context, params types.Params) {
	bz := k.cdc.MustMarshal(&params)
	k.GetStore(ctx, types.StoreModulePrefix).Set([]byte(types.ParamsKey), bz)
}

// GetBridgeContractAddress returns the bridge contract address on ETH
func (k Keeper) GetBridgeContractAddress(ctx context.Context) (*types.EthAddress, error) {
	a := k.GetParams(ctx).BridgeEthereumAddress
	addr, err := types.NewEthAddress(a)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "found invalid bridge contract address in store: %v", a)
	}
	return addr, nil
}

// GetBridgeChainID returns the chain id of the ETH chain we are running against
func (k Keeper) GetBridgeChainID(ctx context.Context) uint64 {
	a := k.GetParams(ctx).BridgeChainId
	return a
}

// Logger returns a module-specific Logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return liblog.FromSDKLogger(sdkCtx.Logger()).With("module", fmt.Sprintf("x/%s", types.ModuleName))
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

// Returns the module store, prefixed with the passed chainReferenceID.
func (k Keeper) GetStore(ctx context.Context, chainReferenceID string) storetypes.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return prefix.NewStore(k.storeGetter.Store(sdkCtx), []byte(chainReferenceID))
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

func (k Keeper) BridgeTax(ctx context.Context) (*types.BridgeTax, error) {
	var tax types.BridgeTax
	val := k.GetStore(ctx, types.StoreModulePrefix).Get(types.BridgeTaxKey)

	if len(val) == 0 {
		// We have no bridge tax settings
		return nil, keeperutil.ErrNotFound.Format(&tax, hex.EncodeToString(types.BridgeTaxKey))
	}

	if err := k.cdc.Unmarshal(val, &tax); err != nil {
		return nil, err
	}

	return &tax, nil
}

func (k Keeper) SetBridgeTax(ctx context.Context, tax *types.BridgeTax) error {
	taxRate, ok := new(big.Rat).SetString(tax.Rate)
	// Sanity check the tax value. Must be between >= 0
	if !ok || taxRate.Sign() < 0 {
		return fmt.Errorf("invalid tax rate value: %s", tax.Rate)
	}

	val, err := k.cdc.Marshal(tax)
	if err != nil {
		return err
	}

	k.GetStore(ctx, types.StoreModulePrefix).Set(types.BridgeTaxKey, val)

	return nil
}

// Calculate the applicable bridge tax amount by checking the current bridge tax
// settings, as well as the transfer sender address and denomination.
// Returns the total amount of tax on the transfer, truncated to integer.
func (k Keeper) bridgeTaxAmount(
	ctx context.Context,
	sender sdk.AccAddress,
	amount sdk.Coin,
) (math.Int, error) {
	bridgeTax, err := k.BridgeTax(ctx)
	if err != nil {
		if errors.Is(err, keeperutil.ErrNotFound) {
			// The bridge tax hasn't been set by governance vote
			return math.ZeroInt(), nil
		}

		return math.ZeroInt(), err
	}

	bRate, _ := new(big.Rat).SetString(bridgeTax.Rate)
	if bRate.Sign() == 0 {
		// The bridge tax is set to zero
		return math.ZeroInt(), nil
	}

	if slices.Contains(bridgeTax.ExcludedTokens, amount.Denom) {
		// This token is excluded, so no tax applies
		return math.ZeroInt(), nil
	}

	for _, addr := range bridgeTax.ExemptAddresses {
		if sender.Equals(addr) {
			// The sender is exempt from bridge tax
			return math.ZeroInt(), nil
		}
	}

	num := math.NewIntFromBigInt(bRate.Num())
	denom := math.NewIntFromBigInt(bRate.Denom())

	return amount.Amount.Mul(num).Quo(denom), nil
}

func (k Keeper) AllBridgeTransferLimits(
	ctx context.Context,
) ([]*types.BridgeTransferLimit, error) {
	st := k.GetStore(ctx, types.BridgeTransferLimitPrefix)
	_, all, err := keeperutil.IterAll[*types.BridgeTransferLimit](st, k.cdc)
	return all, err
}

func (k Keeper) BridgeTransferLimit(
	ctx context.Context,
	denom string,
) (*types.BridgeTransferLimit, error) {
	st := k.GetStore(ctx, types.BridgeTransferLimitPrefix)
	return keeperutil.Load[*types.BridgeTransferLimit](st, k.cdc, []byte(denom))
}

func (k Keeper) SetBridgeTransferLimit(
	ctx context.Context,
	limit *types.BridgeTransferLimit,
) error {
	st := k.GetStore(ctx, types.BridgeTransferLimitPrefix)
	return keeperutil.Save(st, k.cdc, []byte(limit.Token), limit)
}
