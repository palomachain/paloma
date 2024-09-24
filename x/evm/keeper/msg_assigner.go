package keeper

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/evm/types"
	metrixtypes "github.com/palomachain/paloma/x/metrix/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

// topValidatorPoolSize is the number of validators considered when picking
const topValidatorPoolSize = 5

type msgAssigner struct {
	ValsetKeeper   types.ValsetKeeper
	metrixKeeper   types.MetrixKeeper
	treasuryKeeper types.TreasuryKeeper
	logProvider    func(ctx context.Context) liblog.Logr
	scores         scoreSnapshot
}

func newMsgAssigner(v types.ValsetKeeper, m types.MetrixKeeper, t types.TreasuryKeeper, lp func(ctx context.Context) liblog.Logr) msgAssigner {
	return msgAssigner{
		ValsetKeeper:   v,
		metrixKeeper:   m,
		treasuryKeeper: t,
		logProvider:    lp,
		scores:         scoreSnapshot{blockHeight: -1},
	}
}

type ValidatorInfo struct {
	Fee           math.LegacyDec
	Uptime        math.LegacyDec
	SuccessRate   math.LegacyDec
	ExecutionTime math.LegacyDec
	FeatureSet    math.LegacyDec
}

func (ma msgAssigner) PickValidatorForMessage(ctx context.Context, weights *types.RelayWeights, chainReferenceID string, req *xchain.JobRequirements) (string, error) {
	currentSnapshot, err := ma.ValsetKeeper.GetCurrentSnapshot(ctx)
	if err != nil {
		return "", err
	}
	// GetCurrentSnapshot returns nil, nil when no shapshot exists.  We have to catch that and turn it into an error
	if currentSnapshot == nil {
		return "", errors.New("no snapshot found")
	}

	valLUT := make(map[string]valsettypes.Validator)
	for _, val := range currentSnapshot.Validators {
		valLUT[val.Address.String()] = val
	}

	decWeights, err := weights.ValueOrDefault().DecValues()
	if err != nil {
		return "", fmt.Errorf("error converting relay weights: %w", err)
	}

	ma.scores, err = ma.getSnapshotForRound(ctx, ma.scores, chainReferenceID, currentSnapshot.Validators, decWeights)
	if err != nil {
		return "", fmt.Errorf("error getting snapshot for round: %w", err)
	}

	assignableValidators := filterValidatorsForJob(ma.scores, valLUT, chainReferenceID, req)
	if len(assignableValidators) == 0 {
		return "", errors.New("no assignable validators for message")
	}

	ts := sdk.UnwrapSDKContext(ctx).BlockTime().Unix()

	// Use block timestamp to pick a random validator from the top validators
	winnerIdx := ts % int64(min(len(assignableValidators), topValidatorPoolSize))

	winner := assignableValidators[winnerIdx].address
	ma.scores = removeWinnerFromSnapshot(ma.scores, winner)
	return winner, nil
}

func (ma msgAssigner) buildValidatorsInfos(ctx context.Context, chainReferenceID string, validators []valsettypes.Validator) (map[string]ValidatorInfo, error) {
	validatorsInfos := make(map[string]ValidatorInfo)
	perf, err := ma.metrixKeeper.Validators(ctx, nil)
	if err != nil {
		return nil, err
	}

	perfLkup := make(map[string]metrixtypes.ValidatorMetrics)
	for _, v := range perf.GetValMetrics() {
		perfLkup[v.ValAddress] = v
	}

	fees, err := ma.treasuryKeeper.GetRelayerFeesByChainReferenceID(ctx, chainReferenceID)
	if err != nil {
		return nil, err
	}

	for _, validator := range validators {
		validatorAddress := validator.Address.String()
		p, ok := perfLkup[validatorAddress]
		if !ok {
			ma.logger(ctx).WithValidator(validatorAddress).Debug("Validator not found in performance data. Skipping.")
			continue
		}
		fee, ok := fees[validatorAddress]
		if !ok {
			ma.logger(ctx).WithValidator(validatorAddress).Debug("Validator not found in fee data. Skipping.")
			continue
		}
		validatorsInfos[validatorAddress] = ValidatorInfo{
			Fee:           fee,
			Uptime:        p.Uptime,
			SuccessRate:   p.SuccessRate,
			ExecutionTime: p.ExecutionTime.ToLegacyDec(),
			FeatureSet:    p.FeatureSet,
		}
	}

	return validatorsInfos, nil
}

func (ma msgAssigner) logger(ctx context.Context) liblog.Logr {
	return ma.logProvider(ctx).WithComponent("msgAttester")
}

func scoreValue(max, min, val math.LegacyDec, reverse bool) math.LegacyDec {
	// Short circuit.  If all fields are the same, score zero to avoid dividing by zero
	if max.Equal(min) {
		return math.LegacyZeroDec()
	}

	// score := (val - min) / (max - min)
	score := val.Sub(min).Quo(max.Sub(min))

	if reverse {
		return math.LegacyNewDec(1).Sub(score)
	} else {
		return score
	}
}

// We need to disinguish an unset bottom window value from a set bottom window value of 0
// So we use pointers here instead.
type perfDataWindow struct {
	maxFee           *math.LegacyDec
	minFee           *math.LegacyDec
	maxUptime        *math.LegacyDec
	minUptime        *math.LegacyDec
	maxSuccessRate   *math.LegacyDec
	minSuccessRate   *math.LegacyDec
	maxExecutionTime *math.LegacyDec
	minExecutionTime *math.LegacyDec
	maxFeatureSet    *math.LegacyDec
	minFeatureSet    *math.LegacyDec
}

func (pdw perfDataWindow) assert() bool {
	return pdw.maxFee != nil && pdw.minFee != nil &&
		pdw.maxUptime != nil && pdw.minUptime != nil &&
		pdw.maxSuccessRate != nil && pdw.minSuccessRate != nil &&
		pdw.maxExecutionTime != nil && pdw.minExecutionTime != nil &&
		pdw.maxFeatureSet != nil && pdw.minFeatureSet != nil
}

func probeMin(a math.LegacyDec, b *math.LegacyDec) *math.LegacyDec {
	if b == nil {
		return clampToZero(a)
	}
	if a.LT(*b) {
		return clampToZero(a)
	}
	return b
}

func probeMax(a math.LegacyDec, b *math.LegacyDec) *math.LegacyDec {
	if b == nil {
		return &a
	}
	if a.GT(*b) {
		return &a
	}
	return b
}

func clampToZero(a math.LegacyDec) *math.LegacyDec {
	if a.LT(math.LegacyZeroDec()) {
		return &[]math.LegacyDec{math.LegacyZeroDec()}[0]
	}

	return &a
}

type validatorScore struct {
	address string
	score   math.LegacyDec
}

type scoreSnapshot struct {
	scores      []validatorScore
	blockHeight int64
}

func (ma *msgAssigner) getSnapshotForRound(ctx context.Context, snapshot scoreSnapshot, chainReferenceID string, validators []valsettypes.Validator, weights types.RelayWeightDec) (scoreSnapshot, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if sdkCtx.BlockHeight() == snapshot.blockHeight {
		return snapshot, nil
	}

	validatorsInfos, err := ma.buildValidatorsInfos(ctx, chainReferenceID, validators)
	if err != nil {
		return scoreSnapshot{}, fmt.Errorf("error building validators infos: %w", err)
	}

	if len(validatorsInfos) < 1 {
		return scoreSnapshot{}, errors.New("no validators eligible for assignment")
	}

	return rankValidators(ctx, validatorsInfos, weights)
}

func rankValidators(ctx context.Context, validatorsInfos map[string]ValidatorInfo, relayWeights types.RelayWeightDec) (scoreSnapshot, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	ranked := make([]validatorScore, 0, len(validatorsInfos))

	numValidators := len(validatorsInfos)

	infosSlice := make([]ValidatorInfo, 0, numValidators)
	for _, info := range validatorsInfos {
		infosSlice = append(infosSlice, info)
	}

	pdw := slice.Reduce[perfDataWindow, ValidatorInfo](infosSlice, func(window perfDataWindow, val ValidatorInfo) perfDataWindow {
		window.minFee = probeMin(val.Fee, window.minFee)
		window.maxFee = probeMax(val.Fee, window.maxFee)
		window.minUptime = probeMin(val.Uptime, window.minUptime)
		window.maxUptime = probeMax(val.Uptime, window.maxUptime)
		window.minSuccessRate = probeMin(val.SuccessRate, window.minSuccessRate)
		window.maxSuccessRate = probeMax(val.SuccessRate, window.maxSuccessRate)
		window.minExecutionTime = probeMin(val.ExecutionTime, window.minExecutionTime)
		window.maxExecutionTime = probeMax(val.ExecutionTime, window.maxExecutionTime)
		window.minFeatureSet = probeMin(val.FeatureSet, window.minFeatureSet)
		window.maxFeatureSet = probeMax(val.FeatureSet, window.maxFeatureSet)
		return window
	})

	if !pdw.assert() {
		// If we can't assert the metadata, we can't rank the validators.
		return scoreSnapshot{}, errors.New("performance data window assertion failed")
	}

	for valAddr, info := range validatorsInfos {
		scores := map[string]math.LegacyDec{
			"fee":           scoreValue(*pdw.maxFee, *pdw.minFee, info.Fee, true),
			"uptime":        scoreValue(*pdw.maxUptime, *pdw.minUptime, info.Uptime, false),
			"successRate":   scoreValue(*pdw.maxSuccessRate, *pdw.minSuccessRate, info.SuccessRate, false),
			"executionTime": scoreValue(*pdw.maxExecutionTime, *pdw.minExecutionTime, info.ExecutionTime, true),
			"featureSet":    scoreValue(*pdw.maxFeatureSet, *pdw.minFeatureSet, info.FeatureSet, false),
		}

		score := (scores["fee"].Mul(relayWeights.Fee)).Add(
			(scores["uptime"].Mul(relayWeights.Uptime))).Add(
			(scores["successRate"].Mul(relayWeights.SuccessRate))).Add(
			(scores["executionTime"].Mul(relayWeights.ExecutionTime))).Add(
			(scores["featureSet"].Mul(relayWeights.FeatureSet)))

		ranked = append(ranked, validatorScore{address: valAddr, score: score})
	}

	// Sort by score
	slices.SortStableFunc(ranked, func(a, b validatorScore) int {
		if a.score.GT(b.score) {
			return -1
		} else if a.score.LT(b.score) {
			return 1
		}

		// If both validators have the same score, sort them by address
		return strings.Compare(a.address, b.address)
	})
	return scoreSnapshot{
		scores:      ranked,
		blockHeight: sdkCtx.BlockHeight(),
	}, nil
}

// filterValidatorsForJob takes a sorted list of validators and filters them based on the job requirements.
func filterValidatorsForJob(scoreSnapshot scoreSnapshot, valLUT map[string]valsettypes.Validator, chainID string, req *xchain.JobRequirements) []validatorScore {
	return slice.Filter(scoreSnapshot.scores, func(score validatorScore) bool {
		// Snapshot validator not found in lookup.
		// This shouldn't really happen ever, but better to catch it just in case.
		val, ok := valLUT[score.address]
		if !ok {
			return false
		}
		for _, v := range val.ExternalChainInfos {
			// Has support for chain. Should always be the case.
			if v.ChainReferenceID != chainID {
				continue
			}

			// Is MEV support required?
			if req == nil || !req.EnforceMEVRelay {
				return true
			}

			for _, t := range v.Traits {
				if req.EnforceMEVRelay && t == valsettypes.PIGEON_TRAIT_MEV {
					return true
				}
			}

			return false
		}

		return false
	})
}

// removeWinnerFromSnapshot removes the winner address from the snapshot
// It's implementation is O(n) because we don't expect the hit to be very late.
func removeWinnerFromSnapshot(snapshot scoreSnapshot, winner string) scoreSnapshot {
	for i, score := range snapshot.scores {
		if score.address == winner {
			snapshot.scores = append(snapshot.scores[:i], snapshot.scores[i+1:]...)
			break
		}
	}
	return snapshot
}
