package libcons

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	sdkmath "cosmossdk.io/math"
	"github.com/VolumeFi/whoops"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/util/liberr"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
)

const ErrConsensusNotAchieved = liberr.Error("evm: consensus not achieved")

type SnapshotProvider func(context.Context) (*valsettypes.Snapshot, error)

type consensusPower struct {
	runningSum sdkmath.Int
	totalPower sdkmath.Int
}

func (c *consensusPower) setTotal(total sdkmath.Int) {
	c.totalPower = total
}

func (c *consensusPower) add(power sdkmath.Int) {
	var zero sdkmath.Int
	if c.runningSum == zero {
		c.runningSum = sdkmath.NewInt(0)
	}
	c.runningSum = c.runningSum.Add(power)
}

func (c *consensusPower) consensus() bool {
	var zero sdkmath.Int
	if c.runningSum == zero {
		return false
	}
	/*
		sum >= totalPower * 2 / 3
		===
		3 * sum >= totalPower * 2
	*/
	return c.runningSum.Mul(sdkmath.NewInt(3)).GTE(
		c.totalPower.Mul(sdkmath.NewInt(2)),
	)
}

func hashSha256(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

type ConsensusChecker struct {
	p   SnapshotProvider
	cdc codec.BinaryCodec
}

func New(p SnapshotProvider, cdc codec.BinaryCodec) *ConsensusChecker {
	return &ConsensusChecker{
		p:   p,
		cdc: cdc,
	}
}

type Result struct {
	Winner       any
	TotalShares  sdkmath.Int
	TotalVotes   sdkmath.Int
	Distribution map[string]sdkmath.Int
}

func newResult() *Result {
	return &Result{
		TotalShares:  sdkmath.NewInt(0),
		TotalVotes:   sdkmath.NewInt(0),
		Distribution: make(map[string]sdkmath.Int),
	}
}

func (r *Result) totalFromConsensus(c consensusPower) {
	r.TotalVotes = c.runningSum
	r.TotalShares = c.totalPower
}

func (r *Result) addToDistribution(key string, c consensusPower) {
	r.Distribution[key] = c.runningSum
}

func (c ConsensusChecker) VerifyEvidence(ctx sdk.Context, evidences []*consensustypes.Evidence) (*Result, error) {
	result := newResult()
	snapshot, err := c.p(ctx)
	if err != nil {
		return nil, err
	}

	// check if there is enough power to reach the consensus
	// in the best case scenario
	var cp consensusPower
	cp.setTotal(snapshot.TotalShares)

	for _, evidence := range evidences {
		val, found := snapshot.GetValidator(evidence.GetValAddress())
		if !found {
			continue
		}
		cp.add(val.ShareCount)
	}

	result.totalFromConsensus(cp)
	if !cp.consensus() {
		return result, ErrConsensusNotAchieved
	}

	groups := make(map[string]struct {
		evidence   types.Hashable
		validators []sdk.ValAddress
	})

	var g whoops.Group
	for _, evidence := range evidences {
		rawProof := evidence.GetProof()
		var hashable types.Hashable
		err := c.cdc.UnpackAny(rawProof, &hashable)
		if err != nil {
			return nil, err
		}

		bytesToHash, err := hashable.BytesToHash()
		if err != nil {
			return nil, err
		}
		hash := hex.EncodeToString(hashSha256(bytesToHash))
		val := groups[hash]
		if val.evidence == nil {
			val.evidence = hashable
		}
		val.validators = append(val.validators, evidence.ValAddress)
		groups[hash] = val
	}

	// TODO: gas management
	// TODO: punishing validators who misbehave
	// TODO: check for every tx if it seems genuine
	for hash, group := range groups {

		var cp consensusPower
		cp.setTotal(snapshot.TotalShares)

		for _, val := range group.validators {
			snapshotVal, ok := snapshot.GetValidator(val)
			if !ok {
				// strange...
				continue
			}
			cp.add(snapshotVal.ShareCount)
		}

		if cp.consensus() {
			// consensus reached
			result.Winner = group.evidence
			return result, nil
		}

		// TODO: punish other validators that are a part of different groups?
		result.addToDistribution(hash, cp)
	}

	if g.Err() {
		return nil, g
	}

	return result, ErrConsensusNotAchieved
}
