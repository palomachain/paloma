package keeper

import (
	"context"
	"errors"
	"fmt"
	"math/big"

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
	"github.com/palomachain/paloma/v2/util/eventbus"
	keeperutil "github.com/palomachain/paloma/v2/util/keeper"
	"github.com/palomachain/paloma/v2/util/liblog"
	"github.com/palomachain/paloma/v2/x/skyway/types"
)

// Check that our expected keeper types are implemented
var (
	_ types.StakingKeeper      = (*stakingkeeper.Keeper)(nil)
	_ types.SlashingKeeper     = (*slashingkeeper.Keeper)(nil)
	_ types.DistributionKeeper = (*distrkeeper.Keeper)(nil)
)

// Keeper maintains the link to storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	cdc                codec.BinaryCodec // The wire codec for binary encoding/decoding.
	bankKeeper         types.BankKeeper
	StakingKeeper      types.StakingKeeper
	SlashingKeeper     types.SlashingKeeper
	DistKeeper         distrkeeper.Keeper
	accountKeeper      types.AccountKeeper
	ibcTransferKeeper  ibctransferkeeper.Keeper
	EVMKeeper          types.EVMKeeper
	consensusKeeper    types.ConsensusKeeper
	palomaKeeper       types.PalomaKeeper
	ValsetKeeper       types.ValsetKeeper
	tokenFactoryKeeper types.TokenFactoryKeeper
	AddressCodec       address.Codec
	storeGetter        keeperutil.StoreGetter

	AttestationHandler interface {
		Handle(context.Context, types.Attestation, types.EthereumClaim) error
	}
	authority string
}

// NewKeeper returns a new instance of the skyway keeper
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
	palomaKeeper types.PalomaKeeper,
	valsetKeeper types.ValsetKeeper,
	tokenFactoryKeeper types.TokenFactoryKeeper,
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
		EVMKeeper:          evmKeeper,
		consensusKeeper:    consensusKeeper,
		palomaKeeper:       palomaKeeper,
		ValsetKeeper:       valsetKeeper,
		tokenFactoryKeeper: tokenFactoryKeeper,
		storeGetter:        storeGetter,
		AttestationHandler: nil,
		AddressCodec:       valAddressCodec,
	}
	attestationHandler := AttestationHandler{keeper: &k}
	attestationHandler.ValidateMembers()
	k.AttestationHandler = attestationHandler
	k.authority = authority

	eventbus.EVMActivatedChain().Subscribe(
		"skyway-keeper",
		func(ctx context.Context, e eventbus.EVMActivatedChainEvent) error {
			k.setLatestCompassID(ctx, e.ChainReferenceID, string(e.SmartContractUniqueID))
			return k.overrideNonce(ctx, e.ChainReferenceID, 0)
		})

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
// Adding here in skyway keeper as cdc is not available inside endblocker.
func (k Keeper) DeserializeValidatorIterator(vals []byte) stakingtypes.ValAddresses {
	validators := stakingtypes.ValAddresses{
		Addresses: []string{},
	}
	k.cdc.MustUnmarshal(vals, &validators)
	return validators
}

// InvalidSendToRemoteAddress Returns true if the provided address is invalid to send to Ethereum this could be
// for one of several reasons. (1) it is invalid in general like the Zero address, (2)
// it is invalid for a subset of ERC20 addresses. (2) is not yet implemented
// Blocking some addresses is technically motivated, if any ERC20 transfers in a batch fail the entire batch
// becomes impossible to execute.
func (k Keeper) InvalidSendToRemoteAddress(ctx context.Context, addr types.EthAddress, _erc20Addr types.EthAddress) bool {
	return addr == types.ZeroAddress()
}

// Returns the module store, prefixed with the passed chainReferenceID.
func (k Keeper) GetStore(ctx context.Context, chainReferenceID string) storetypes.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return prefix.NewStore(k.storeGetter.Store(sdkCtx), []byte(chainReferenceID))
}

type SkywayStoreGetter struct {
	storeKey storetypes.StoreKey
}

func NewSkywayStoreGetter(storeKey storetypes.StoreKey) SkywayStoreGetter {
	return SkywayStoreGetter{
		storeKey: storeKey,
	}
}

func (gsg SkywayStoreGetter) Store(ctx context.Context) storetypes.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.KVStore(gsg.storeKey)
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) AllBridgeTaxes(
	ctx context.Context,
) ([]*types.BridgeTax, error) {
	st := k.GetStore(ctx, types.BridgeTaxPrefix)
	_, all, err := keeperutil.IterAll[*types.BridgeTax](st, k.cdc)
	return all, err
}

func (k Keeper) BridgeTax(
	ctx context.Context,
	token string,
) (*types.BridgeTax, error) {
	st := k.GetStore(ctx, types.BridgeTaxPrefix)
	return keeperutil.Load[*types.BridgeTax](st, k.cdc, []byte(token))
}

func (k Keeper) SetBridgeTax(ctx context.Context, tax *types.BridgeTax) error {
	if tax.Token == "" {
		return errors.New("empty tax token")
	}

	taxRate, ok := new(big.Rat).SetString(tax.Rate)
	// Sanity check the tax value. Must be >= 0
	if !ok || taxRate.Sign() < 0 {
		return fmt.Errorf("invalid tax rate value: %s", tax.Rate)
	}

	st := k.GetStore(ctx, types.BridgeTaxPrefix)
	return keeperutil.Save(st, k.cdc, []byte(tax.Token), tax)
}

// Calculate the applicable bridge tax amount by checking the current bridge tax
// settings, as well as the transfer sender address and denomination.
// Returns the total amount of tax on the transfer, truncated to integer.
func (k Keeper) bridgeTaxAmount(
	ctx context.Context,
	sender sdk.AccAddress,
	coin sdk.Coin,
) (math.Int, error) {
	bridgeTax, err := k.BridgeTax(ctx, coin.Denom)
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

	for _, addr := range bridgeTax.ExemptAddresses {
		if sender.Equals(addr) {
			// The sender is exempt from bridge tax
			return math.ZeroInt(), nil
		}
	}

	num := math.NewIntFromBigInt(bRate.Num())
	denom := math.NewIntFromBigInt(bRate.Denom())

	return coin.Amount.Mul(num).Quo(denom), nil
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
	token string,
) (*types.BridgeTransferLimit, error) {
	st := k.GetStore(ctx, types.BridgeTransferLimitPrefix)
	return keeperutil.Load[*types.BridgeTransferLimit](st, k.cdc, []byte(token))
}

func (k Keeper) SetBridgeTransferLimit(
	ctx context.Context,
	limit *types.BridgeTransferLimit,
) error {
	if limit.Token == "" {
		return errors.New("empty transfer limit token")
	}

	st := k.GetStore(ctx, types.BridgeTransferLimitPrefix)
	return keeperutil.Save(st, k.cdc, []byte(limit.Token), limit)
}

func (k Keeper) BridgeTransferUsage(
	ctx context.Context,
	token string,
) (*types.BridgeTransferUsage, error) {
	st := k.GetStore(ctx, types.BridgeTransferUsagePrefix)
	return keeperutil.Load[*types.BridgeTransferUsage](st, k.cdc, []byte(token))
}

func (k Keeper) UpdateBridgeTransferUsageWithLimit(
	ctx context.Context,
	sender sdk.AccAddress,
	coin sdk.Coin,
) error {
	limits, err := k.BridgeTransferLimit(ctx, coin.Denom)
	if err != nil {
		if errors.Is(err, keeperutil.ErrNotFound) {
			// If we have no limits set, there's nothing else to do
			return nil
		}

		return err
	}

	for _, addr := range limits.ExemptAddresses {
		if sender.Equals(addr) {
			// The sender is exempt from bridge transfer limits
			return nil
		}
	}

	if limits.LimitPeriod == types.LimitPeriod_NONE {
		// We have data in storage, but it doesn't impose limits
		return nil
	}

	usage, err := k.BridgeTransferUsage(ctx, coin.Denom)
	if err != nil && !errors.Is(err, keeperutil.ErrNotFound) {
		// If we have no usage information, we will set it after this
		return err
	}

	blockHeight := sdk.UnwrapSDKContext(ctx).BlockHeight()

	var newUsage types.BridgeTransferUsage
	if usage == nil || usage.Total.IsNil() ||
		blockHeight-usage.StartBlockHeight >= limits.BlockLimit() {

		// Either there's no usage information, or we're past the limit block
		// range, so we want to overwrite it with recent data
		newUsage = types.BridgeTransferUsage{
			Total:            coin.Amount,
			StartBlockHeight: blockHeight,
		}
	} else {
		// Otherwise, we have an ongoing tally, so we add to it but keep the
		// reference block height
		newUsage = types.BridgeTransferUsage{
			Total:            usage.Total.Add(coin.Amount),
			StartBlockHeight: usage.StartBlockHeight,
		}
	}

	if newUsage.Total.GT(limits.Limit) {
		// Before persisting it, we check the limit
		return fmt.Errorf("limit for bridge transfer reached %v", limits.Limit)
	}

	// Allow the transfer and update the counters

	st := k.GetStore(ctx, types.BridgeTransferUsagePrefix)
	return keeperutil.Save(st, k.cdc, []byte(coin.Denom), &newUsage)
}

func (k Keeper) AllLightNodeSaleContracts(
	ctx context.Context,
) ([]*types.LightNodeSaleContract, error) {
	st := k.GetStore(ctx, types.LightNodeSaleContractsPrefix)
	_, all, err := keeperutil.IterAll[*types.LightNodeSaleContract](st, k.cdc)
	return all, err
}

func (k Keeper) LightNodeSaleContract(
	ctx context.Context,
	chain string,
) (*types.LightNodeSaleContract, error) {
	st := k.GetStore(ctx, types.LightNodeSaleContractsPrefix)
	return keeperutil.Load[*types.LightNodeSaleContract](st, k.cdc, []byte(chain))
}

func (k Keeper) SetAllLighNodeSaleContracts(
	ctx context.Context,
	contracts []*types.LightNodeSaleContract,
) error {
	st := k.GetStore(ctx, types.LightNodeSaleContractsPrefix)

	// Delete all existing contracts
	err := keeperutil.IterAllFnc(st, k.cdc, func(key []byte, _ *types.LightNodeSaleContract) bool {
		st.Delete(key)
		return true
	})
	if err != nil {
		return err
	}

	// Store the new ones
	for _, contract := range contracts {
		err := keeperutil.Save(st, k.cdc, []byte(contract.ChainReferenceId), contract)
		if err != nil {
			return err
		}
	}

	return nil
}

// overrideNonce purposefully circumvents keeper setters and getters, as those perform
// some biased operations and may end up inserting a different value or dropping it altogether.
// it should never be used outside of the EVM activation event handler or a
// proposal.
func (k Keeper) overrideNonce(ctx context.Context, chainReferenceId string, nonce uint64) error {
	cacheCtx, commit := sdk.UnwrapSDKContext(ctx).CacheContext()
	logger := liblog.FromKeeper(ctx, k).WithComponent("nonce-override")

	store := k.GetStore(ctx, chainReferenceId)
	store.Set(types.LastObservedEventNonceKey, types.UInt64Bytes(nonce))

	keys := [][]byte{}
	err := k.IterateValidatorLastEventNonces(cacheCtx, chainReferenceId, func(key []byte, _ uint64) bool {
		keys = append(keys, key)
		return false
	})
	if err != nil {
		logger.WithError(err).Warn("Failed to reset validator skyway nonces")
		return err
	}

	for _, key := range keys {
		prefixStore := prefix.NewStore(store, types.LastEventNonceByValidatorKey)
		prefixStore.Set(key, types.UInt64Bytes(nonce))
	}

	commit()
	logger.Info("Updated last observed nonce successfully")
	return nil
}

// UpdateValidatorNoncesToLatest updates all validator nonces to the last
// observed skyway nonce only if it is higher than the current record
func (k Keeper) UpdateValidatorNoncesToLatest(
	ctx context.Context,
	chainReferenceId string,
) error {
	logger := liblog.FromKeeper(ctx, k).WithComponent("update-validator-nonces-if-higher")

	lastSkywayNonce, err := k.GetLastObservedSkywayNonce(ctx, chainReferenceId)
	if err != nil {
		logger.WithError(err).Warn("Failed to get last observed nonce.")
		return err
	}

	store := k.GetStore(ctx, chainReferenceId)
	prefixStore := prefix.NewStore(store, types.LastEventNonceByValidatorKey)

	err = k.IterateValidatorLastEventNonces(ctx, chainReferenceId, func(key []byte, nonce uint64) bool {
		if lastSkywayNonce > nonce {
			prefixStore.Set(key, types.UInt64Bytes(lastSkywayNonce))
		}

		return false
	})
	if err != nil {
		logger.WithError(err).Warn("Failed to update validator skyway nonces")
		return err
	}

	logger.Info("Updated validator nonces to highest successfully")
	return nil
}
