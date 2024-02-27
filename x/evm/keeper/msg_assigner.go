package keeper

import (
	"context"
	"errors"
	"math"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/evm/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

type MsgAssigner struct {
	ValsetKeeper types.ValsetKeeper
}

func filterAssignableValidators(validators []valsettypes.Validator, chainID string, req *xchain.JobRequirements) []valsettypes.Validator {
	return slice.Filter(validators, func(val valsettypes.Validator) bool {
		for _, v := range val.ExternalChainInfos {
			if v.ChainReferenceID != chainID {
				continue
			}

			if req == nil || !req.EnforceMEVRelay {
				return true
			}

			for _, t := range v.Traits {
				if req.EnforceMEVRelay && t == valsettypes.PIGEON_TRAIT_MEV {
					return true
				}
			}
		}

		return false
	})
}

func getValidatorsFees(validators []valsettypes.Validator) map[string]float64 {
	validatorsFees := make(map[string]float64, len(validators))
	for _, validator := range validators {
		// Placeholder until we start tracking validator fees
		validatorsFees[validator.Address.String()] = 0.05
	}
	return validatorsFees
}

func getValidatorsUptime(validators []valsettypes.Validator) map[string]float64 {
	validatorsUptime := make(map[string]float64, len(validators))
	for _, validator := range validators {
		// Placeholder until we start tracking validator uptime
		validatorsUptime[validator.Address.String()] = 1.0
	}
	return validatorsUptime
}

func getValidatorsSuccessRate(validators []valsettypes.Validator) map[string]float64 {
	validatorsSuccessRate := make(map[string]float64, len(validators))
	for _, validator := range validators {
		// Placeholder until we start tracking validator success rate
		validatorsSuccessRate[validator.Address.String()] = 1.0
	}
	return validatorsSuccessRate
}

func getValidatorsExecutionTime(validators []valsettypes.Validator) map[string]float64 {
	validatorsSuccessRate := make(map[string]float64, len(validators))
	for _, validator := range validators {
		// Placeholder until we start tracking validator executionTime
		validatorsSuccessRate[validator.Address.String()] = 0.5
	}
	return validatorsSuccessRate
}

func getValidatorsFeatureSet(validators []valsettypes.Validator) map[string]float64 {
	validatorsFeatureSet := make(map[string]float64, len(validators))
	for _, validator := range validators {
		// Placeholder until we start tracking validator feature sets
		validatorsFeatureSet[validator.Address.String()] = 0.5
	}
	return validatorsFeatureSet
}

type ValidatorInfo struct {
	Fee           float64
	Uptime        float64
	SuccessRate   float64
	ExecutionTime float64
	FeatureSet    float64
}

func buildValidatorsInfos(validators []valsettypes.Validator) map[string]ValidatorInfo {
	fees := getValidatorsFees(validators)
	uptime := getValidatorsUptime(validators)
	successRate := getValidatorsSuccessRate(validators)
	executionTime := getValidatorsExecutionTime(validators)
	featureSet := getValidatorsFeatureSet(validators)

	validatorsInfos := make(map[string]ValidatorInfo, len(validators))

	for _, validator := range validators {
		validatorAddress := validator.Address.String()
		validatorsInfos[validatorAddress] = ValidatorInfo{
			Fee:           fees[validatorAddress],
			Uptime:        uptime[validatorAddress],
			SuccessRate:   successRate[validatorAddress],
			ExecutionTime: executionTime[validatorAddress],
			FeatureSet:    featureSet[validatorAddress],
		}
	}

	return validatorsInfos
}

func scoreValue(max, min, val float64, reverse bool) float64 {
	// Short circuit.  If all fields are the same, score zero to avoid dividing by zero
	if max == min {
		return 0.0
	}

	score := (val - min) / (max - min)

	if reverse {
		return 1 - score
	} else {
		return score
	}
}

type validatorsMeta struct {
	maxFee           float64
	minFee           float64
	maxUptime        float64
	minUptime        float64
	maxSuccessRate   float64
	minSuccessRate   float64
	maxExecutionTime float64
	minExecutionTime float64
	maxFeatureSet    float64
	minFeatureSet    float64
}

func rankValidators(validatorsInfos map[string]ValidatorInfo, relayWeights types.RelayWeightsFloat64) map[string]int {
	ranked := make(map[string]int)

	numValidators := len(validatorsInfos)

	infosSlice := make([]ValidatorInfo, numValidators)
	for _, info := range validatorsInfos {
		infosSlice = append(infosSlice, info)
	}

	validatorsMetadata := slice.Reduce[validatorsMeta, ValidatorInfo](infosSlice, func(acc validatorsMeta, val ValidatorInfo) validatorsMeta {
		if val.Fee < acc.minFee || acc.minFee == 0 {
			acc.minFee = val.Fee
		}
		if val.Fee > acc.maxFee {
			acc.maxFee = val.Fee
		}

		if val.Uptime < acc.minUptime || acc.minUptime == 0 {
			acc.minUptime = val.Uptime
		}
		if val.Uptime > acc.maxUptime {
			acc.maxUptime = val.Uptime
		}

		if val.SuccessRate < acc.minSuccessRate || acc.minSuccessRate == 0 {
			acc.minSuccessRate = val.SuccessRate
		}
		if val.SuccessRate > acc.maxSuccessRate {
			acc.maxSuccessRate = val.SuccessRate
		}

		if val.ExecutionTime < acc.minExecutionTime || acc.minExecutionTime == 0 {
			acc.minExecutionTime = val.ExecutionTime
		}
		if val.ExecutionTime > acc.maxExecutionTime {
			acc.maxExecutionTime = val.ExecutionTime
		}

		if val.FeatureSet < acc.minFeatureSet || acc.minFeatureSet == 0 {
			acc.minFeatureSet = val.FeatureSet
		}
		if val.FeatureSet > acc.maxFeatureSet {
			acc.maxFeatureSet = val.FeatureSet
		}

		return acc
	})

	for valAddr, info := range validatorsInfos {
		scores := map[string]float64{
			"fee":           scoreValue(validatorsMetadata.maxFee, validatorsMetadata.minFee, info.Fee, true),
			"uptime":        scoreValue(validatorsMetadata.maxUptime, validatorsMetadata.minUptime, info.Uptime, false),
			"successRate":   scoreValue(validatorsMetadata.maxSuccessRate, validatorsMetadata.minSuccessRate, info.SuccessRate, false),
			"executionTime": scoreValue(validatorsMetadata.maxExecutionTime, validatorsMetadata.minExecutionTime, info.ExecutionTime, true),
			"featureSet":    scoreValue(validatorsMetadata.maxFeatureSet, validatorsMetadata.minFeatureSet, info.FeatureSet, false),
		}

		score := (scores["fee"] * relayWeights.Fee) +
			(scores["uptime"] * relayWeights.Uptime) +
			(scores["successRate"] * relayWeights.SuccessRate) +
			(scores["executionTime"] * relayWeights.ExecutionTime) +
			(scores["featureSet"] * relayWeights.FeatureSet)

		ranked[valAddr] = int(math.Round(score * 10))
	}

	return ranked
}

func pickValidator(ctx context.Context, validatorsInfos map[string]ValidatorInfo, weights types.RelayWeightsFloat64) string {
	scored := rankValidators(validatorsInfos, weights)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	highScore := 0
	var highScorers []string

	// Get the validators with the highest score.
	for addr, score := range scored {
		switch {
		case score > highScore:
			highScore = score
			highScorers = []string{addr}
		case score == highScore:
			highScorers = append(highScorers, addr)
		}
	}

	// All else equal, grab one of our high scorers, but not always the same one
	sort.Strings(highScorers)
	return highScorers[int(sdkCtx.BlockHeight())%len(highScorers)]
}

func (ma MsgAssigner) PickValidatorForMessage(ctx context.Context, weights *types.RelayWeights, chainID string, req *xchain.JobRequirements) (string, error) {
	currentSnapshot, err := ma.ValsetKeeper.GetCurrentSnapshot(ctx)
	if err != nil {
		return "", err
	}

	// For some reason, GetCurrentSnapshot returns nil, nil when no shapshot exists.  We have to catch that and turn it into an error
	if currentSnapshot == nil {
		return "", errors.New("no snapshot found")
	}

	assignableValidators := filterAssignableValidators(currentSnapshot.Validators, chainID, req)
	if len(assignableValidators) == 0 {
		return "", errors.New("no assignable validators for message")
	}

	validatorsInfos := buildValidatorsInfos(assignableValidators)

	if weights == nil {
		weights = &types.RelayWeights{
			ExecutionTime: "1.0",
			SuccessRate:   "1.0",
			Uptime:        "1.0",
			Fee:           "1.0",
			FeatureSet:    "1.0",
		}
	}
	relayWeights := weights.Float64Values()

	return pickValidator(ctx, validatorsInfos, relayWeights), nil
}
