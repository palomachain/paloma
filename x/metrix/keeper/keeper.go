package keeper

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"

	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/util/palomath"
	"github.com/palomachain/paloma/x/metrix/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

const (
	// cRecordHistoryCap specifies the amount of messages within the
	// scoring window to consider during performance scoring.
	// If more messages are processed during the window, the data set
	// will roll over and remove the oldest entries.
	cRecordHistoryCap int = 100

	// cRecordHistoryScoringWindow specifies the window of messages to
	// consider when scoring. Messages older than this window will
	// be purged from the data set.
	cRecordHistoryScoringWindow uint64 = 1000

	// cSuccessRateIncrement specifies the increment applied to the
	// message relay success rate for every successful attempt.
	cSuccessRateIncrement = 0.01

	// cSuccessRateDecrement specifies the decrement applied to the
	// message relay success rate for every failed attempt.
	cSuccessRateDecrement = 0.02
)

var _ valsettypes.OnSnapshotBuiltListener = &Keeper{}

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		paramstore paramtypes.Subspace
		slashing   types.SlashingKeeper
		staking    types.StakingKeeper

		metrics           *keeperutil.KVStoreWrapper[*types.ValidatorMetrics]
		history           *keeperutil.KVStoreWrapper[*types.ValidatorHistory]
		messageNonceCache *keeperutil.KVStoreWrapper[*types.HistoricRelayData]
		AddressCodec      address.Codec
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey corestore.KVStoreService,
	ps paramtypes.Subspace,
	slashing types.SlashingKeeper,
	staking types.StakingKeeper,
	addressCodec address.Codec,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:               cdc,
		paramstore:        ps,
		slashing:          slashing,
		staking:           staking,
		metrics:           keeperutil.NewKvStoreWrapper[*types.ValidatorMetrics](storeFactory(storeKey, types.MetricsStorePrefix), cdc),
		history:           keeperutil.NewKvStoreWrapper[*types.ValidatorHistory](storeFactory(storeKey, types.HistoryStorePrefix), cdc),
		messageNonceCache: keeperutil.NewKvStoreWrapper[*types.HistoricRelayData](storeFactory(storeKey, types.MessageNonceCacheStorePrefix), cdc),
		AddressCodec:      addressCodec,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetValidatorMetrics returns the metrics for a validator. Will return nil if no metrics found
// for the given validator address.
func (k Keeper) GetValidatorMetrics(ctx context.Context, address sdk.ValAddress) (*types.ValidatorMetrics, error) {
	return getFromStore(ctx, k.metrics, address)
}

// GetValidatorHistory returns the historic relay data for a validator. Will return nil if no history found
// for the given validator address.
func (k Keeper) GetValidatorHistory(ctx context.Context, address sdk.ValAddress) (*types.ValidatorHistory, error) {
	return getFromStore(ctx, k.history, address)
}

// GetValidatorMetrics returns the metrics for a validator. Will return nil if no metrics found
// for the given validator address.
func (k Keeper) GetMessageNonceCache(ctx context.Context) (*types.HistoricRelayData, error) {
	return getFromStore(ctx, k.messageNonceCache, types.MessageNonceCacheKey)
}

// OnConsensusMessageAttested implements types.OnConsensusMessageAttestedListener.
func (k Keeper) OnConsensusMessageAttested(goCtx context.Context, e types.MessageAttestedEvent) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get record from store
	logger := liblog.FromSDKLogger(k.Logger(ctx)).
		WithComponent("metrix.OnConsensusMessageAttested").
		WithFields("event", e)

	logger.Debug("Running OnConsensusMessageAttested")

	if e.HandledAtBlockHeight.LT(e.AssignedAtBlockHeight) ||
		e.HandledAtBlockHeight.GT(math.NewInt(ctx.BlockHeight())) {
		logger.Error("Skipping message with invalid block heights.")
		return
	}

	record, err := k.GetValidatorHistory(ctx, e.Assignee)
	if err != nil {
		logger.WithError(err).Error("Failed to get history from store")
		return
	}
	if record == nil {
		logger.Debug("No record found, creating new record...")
		record = &types.ValidatorHistory{
			ValAddress: e.Assignee.String(),
			Records:    make([]types.HistoricRelayData, 0, cRecordHistoryCap),
		}
	}

	// Roll over data set if cap exceeded by popping the
	// oldest entry.
	if len(record.Records) >= cRecordHistoryCap {
		logger.Debug("Record capacity reached, rolling over ...")
		record.Records = record.Records[1:]
	}

	// Append latest record
	record.Records = append(record.Records, types.HistoricRelayData{
		MessageId:              e.MessageID,
		Success:                e.WasRelayedSuccessfully,
		ExecutionSpeedInBlocks: e.HandledAtBlockHeight.Sub(e.AssignedAtBlockHeight).Uint64(),
	})

	if err := k.history.Set(ctx, e.Assignee, record); err != nil {
		logger.WithError(err).Error("Failed to write history to store")
		return
	}

	cache, err := k.GetMessageNonceCache(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get message nonce cache from store.")
		return
	}
	if cache == nil || cache.MessageId < e.MessageID {
		logger.Debug("Updating message nonce cache.")
		cache = &types.HistoricRelayData{
			MessageId: e.MessageID,
		}
	}

	if err := k.messageNonceCache.Set(ctx, types.MessageNonceCacheKey, cache); err != nil {
		logger.WithError(err).Error("Failed to persist message nonce cache into store.")
		return
	}

	logger.Debug("OnConsensusMessageAttested finished!")
}

// OnSnapshotBuilt implements types.OnSnapshotBuiltListener.
func (k Keeper) OnSnapshotBuilt(ctx context.Context, snapshot *valsettypes.Snapshot) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := liblog.FromSDKLogger(k.Logger(sdkCtx)).WithComponent("metrix.OnSnapshotBuilt")
	logger.Debug("Updating snapshot metrics...")

	// Building the feature set is currently only taking MEV support into consideration.
	for _, v := range snapshot.Validators {
		logger := logger.WithValidator(v.GetAddress().String())
		scoreMax := int64(len(v.ExternalChainInfos))
		if scoreMax < 1 {
			logger.Info("Skip updating metrics, no chains found.")
			return
		}
		scoreMax = int64(len(v.ExternalChainInfos))
		if scoreMax < 1 {
			logger.Info("Skip updating metrics, no chains found.")
			return
		}

		matches := func() int64 {
			if v.State != valsettypes.ValidatorState_ACTIVE {
				logger.Info("Skipping inactive validator.")
				return 0
			}

			var matches int64
			for _, c := range v.ExternalChainInfos {
				if slices.Contains(c.Traits, valsettypes.PIGEON_TRAIT_MEV) {
					matches++
				}
			}

			return matches
		}()

		score := palomath.BigIntDiv(big.NewInt(matches), big.NewInt(scoreMax))
		k.updateRecord(sdkCtx, v.GetAddress(), recordPatch{featureSet: &score})
	}
}

func (k *Keeper) PurgeRelayMetrics(ctx context.Context) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := liblog.FromSDKLogger(k.Logger(sdkCtx)).WithComponent("metrix.PurgeRelayMetrics")
	logger.Debug("Running relay metrics purging loop...")

	cache, err := k.GetMessageNonceCache(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get message nonce from cache.")
		return
	}
	if cache == nil {
		logger.WithError(err).Info("Skipping metrics relay purge, empty cache.")
		return
	}

	if cache.MessageId <= cRecordHistoryScoringWindow {
		logger.WithError(err).Info("Skipping metrics relay purge, cached message ID smaller than scoring window.")
		return
	}

	threshold := cache.MessageId - cRecordHistoryScoringWindow
	updates := make(map[string]types.ValidatorHistory)
	logger = logger.WithFields("threshold", threshold)

	// Purge records outside scoring window
	if err := k.history.Iterate(sdkCtx, func(key []byte, history *types.ValidatorHistory) bool {
		logger := logger.WithValidator(history.ValAddress)
		logger.Debug("Puring records for validator.")

		purged := history.Records

		func() {
			for i, v := range history.Records {
				if v.MessageId >= threshold {
					// from here on, messages are within the window
					logger.Debug(fmt.Sprintf("Puring %d records.", i))
					purged = history.Records[i:]
					return
				}

				// looks like we're purging everything
				logger.Debug("Purging entire history.")
				purged = nil
			}
		}()

		if purged == nil || len(purged) != len(history.Records) {
			updates[string(key)] = types.ValidatorHistory{
				ValAddress: history.ValAddress,
				Records:    purged,
			}
		}

		return true
	}); err != nil {
		logger.WithError(err).Info("Iterating validator history failed.")
		return
	}

	logger.Debug("Executing purge...")
	for key, val := range updates {
		memcpy := val
		if err := k.history.Set(sdkCtx, sdk.ValAddress([]byte(key)), &memcpy); err != nil {
			logger.WithError(err).WithValidator(val.ValAddress).Error("Failed to purge history")
		}
	}
	logger.Debug("Purge finished!")
}

func (k *Keeper) UpdateRelayMetrics(ctx context.Context) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	logger := liblog.FromSDKLogger(k.Logger(sdkCtx)).WithComponent("metrix.UpdateRelayMetrics")
	logger.Debug("Running relay metrics update loop.")

	if err := k.history.Iterate(sdkCtx, func(key []byte, history *types.ValidatorHistory) bool {
		logger := logger.WithValidator(history.ValAddress)
		valAddress := sdk.ValAddress(key)

		metrics, err := k.GetValidatorMetrics(ctx, valAddress)
		if err != nil {
			logger.WithError(err).Error("Failed to get history from store.")
			return true
		}

		successRateScore := 0.5
		executionTimes := make([]uint64, len(history.Records))

		for i, v := range history.Records {
			executionTimes[i] = v.ExecutionSpeedInBlocks
			if v.Success {
				successRateScore += cSuccessRateIncrement
			} else {
				successRateScore -= cSuccessRateDecrement
			}
		}

		successRateScore = palomath.Clamp(successRateScore, 0, 1)
		medianExecutionTime := palomath.Median(executionTimes)

		logger = logger.WithFields("success-rate", successRateScore, "execution-time", medianExecutionTime)
		decSuccessRateScore := palomath.LegacyDecFromFloat64(successRateScore)
		decMedianExecutionTime := math.NewIntFromUint64(medianExecutionTime)

		logger.Debug("Calculated relay  metrics.")
		if metrics == nil ||
			!metrics.SuccessRate.Equal(decSuccessRateScore) ||
			!metrics.ExecutionTime.Equal(decMedianExecutionTime) {
			logger.Debug("Relay metrics changed. Updating validator metrics.")
			k.updateRecord(ctx, valAddress, recordPatch{
				executionTime: &decMedianExecutionTime,
				successRate:   &decSuccessRateScore,
			})
		}

		return true
	}); err != nil {
		logger.WithError(err).Info("Iterating validator history failed.")
		return
	}

	logger.Debug("Updating relay metrics finished!")
}

func (k *Keeper) UpdateUptime(ctx context.Context) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	window, err := k.slashing.SignedBlocksWindow(ctx)
	if err != nil {
		liblog.FromSDKLogger(sdkCtx.Logger()).WithError(err).Error("Error while getting the SignedBlocksWindow")
	}
	logger := liblog.FromSDKLogger(k.Logger(sdkCtx)).WithComponent("metrix.UpdateUptime").WithFields("signed-blocks-window", window)
	logger.Debug("Running uptime update loop.")

	jailed := make(map[string]struct{})
	err = k.staking.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) bool {
		bz, _ := keeperutil.ValAddressFromBech32(k.AddressCodec, val.GetOperator())
		valAddr, _ := k.AddressCodec.BytesToString(bz)
		if val.IsJailed() {
			jailed[valAddr] = struct{}{}
		}
		return false
	})
	if err != nil {
		liblog.FromSDKLogger(sdkCtx.Logger()).WithError(err).Error("Error while IteratingValidators")
	}

	err = k.slashing.IterateValidatorSigningInfos(ctx, func(consAddr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		logger := liblog.FromKeeper(ctx, k).
			WithComponent("metrix.UpdateUptime").
			WithFields("signed-blocks-window", window).
			WithFields("validator-conspub", info.GetAddress())
		val, err := k.staking.GetValidatorByConsAddr(ctx, consAddr)
		if err != nil {
			// SigningInfos might hold infos of validators not staking or bonded
			// This is expected.
			// TODO: A better solution would be to invert the loop and look up signing infos
			// while iterating over vaidators with stake.
			return false
		}
		bz, _ := keeperutil.ValAddressFromBech32(k.AddressCodec, val.GetOperator())
		valAddr, _ := k.AddressCodec.BytesToString(bz)
		uptime := math.LegacyNewDec(0)
		_, isJailed := jailed[valAddr]
		if !isJailed {
			uptime = calculateUptime(window, info.MissedBlocksCounter)
		}
		logger.WithValidator(valAddr).
			WithFields(
				"missed-blocks-counter", info.MissedBlocksCounter,
				"uptime", uptime,
				"is-jailed", isJailed).
			Debug("Calculated uptime, updating record.")
		k.updateRecord(ctx, bz, recordPatch{uptime: &uptime})
		return false
	})
	if err != nil {
		liblog.FromSDKLogger(sdkCtx.Logger()).WithError(err).Error("Error while IterateValidatorSigningInfos")
	}
	logger.Debug("Updating uptime finished!")
}

type recordPatch struct {
	uptime        *math.LegacyDec
	successRate   *math.LegacyDec
	executionTime *math.Int
	fee           *math.Int
	featureSet    *math.LegacyDec
}

func (p recordPatch) apply(data types.ValidatorMetrics) *types.ValidatorMetrics {
	if p.uptime != nil {
		data.Uptime = *p.uptime
	}

	if p.successRate != nil {
		data.SuccessRate = *p.successRate
	}

	if p.executionTime != nil {
		data.ExecutionTime = *p.executionTime
	}

	if p.fee != nil {
		data.Fee = *p.fee
	}

	if p.featureSet != nil {
		data.FeatureSet = *p.featureSet
	}

	return &data
}

func (k Keeper) updateRecord(ctx context.Context, valAddr sdk.ValAddress, patch recordPatch) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := k.tryUpdateRecord(ctx, valAddr, patch); err != nil {
		liblog.FromSDKLogger(k.Logger(sdkCtx)).
			WithError(err).
			WithComponent("metrix.updateRecord").
			WithValidator(valAddr.String()).
			Error("Failed to update metrics record.")
	}
}

func (k Keeper) tryUpdateRecord(ctx context.Context, valAddr sdk.ValAddress, patch recordPatch) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	record, err := k.GetValidatorMetrics(ctx, valAddr)
	if err != nil {
		return fmt.Errorf("update record: %w", err)
	}

	if record == nil {
		record = &types.ValidatorMetrics{
			ValAddress:    valAddr.String(),
			Uptime:        palomath.LegacyDecFromFloat64(1.0),
			SuccessRate:   palomath.LegacyDecFromFloat64(0.5),
			ExecutionTime: math.ZeroInt(),
			Fee:           math.ZeroInt(),
			FeatureSet:    math.LegacyZeroDec(),
		}
	}

	patched := patch.apply(*record)

	// Do not override if no changes
	if record.Equal(patched) {
		return nil
	}

	err = k.metrics.Set(sdkCtx, valAddr, patched)
	if err != nil {
		return fmt.Errorf("update record: %w", err)
	}

	return nil
}

func storeFactory(storeKey corestore.KVStoreService, p string) func(ctx context.Context) storetypes.KVStore {
	return func(ctx context.Context) storetypes.KVStore {
		s := runtime.KVStoreAdapter(storeKey.OpenKVStore(ctx))
		return prefix.NewStore(s, types.KeyPrefix(p))
	}
}

func calculateUptime(window, missed int64) math.LegacyDec {
	if window < 1 || missed < 0 || missed > window {
		return math.LegacyNewDec(0)
	}

	// return (window - missed) / window
	w := big.NewInt(window)
	m := big.NewInt(missed)

	diff := big.NewInt(0).Sub(w, m)
	return palomath.BigIntDiv(diff, w)
}

func getFromStore[T codec.ProtoMarshaler](ctx context.Context, store *keeperutil.KVStoreWrapper[T], key keeperutil.Byter) (T, error) {
	var empty T
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	data, err := store.Get(sdkCtx, key)
	if err != nil {
		if !errors.Is(err, keeperutil.ErrNotFound) {
			return empty, fmt.Errorf("get from store: %w", err)
		}

		return empty, nil
	}

	return data, nil
}
