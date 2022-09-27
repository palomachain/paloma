package keeper

import (
	"errors"
	"fmt"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/vizualni/whoops"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/scheduler/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		paramstore paramtypes.Subspace

		ider keeperutil.IDGenerator

		Chains map[xchain.Type]xchain.Bridge
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	chains []xchain.Bridge,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	cm := slice.MustMakeMapKeys(chains, func(c xchain.Bridge) xchain.Type {
		return c.XChainType()
	})

	k := &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		Chains:     cm,
	}

	k.ider = keeperutil.NewIDGenerator(k, nil)

	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// store returns default store for this keeper!
func (k Keeper) Store(ctx sdk.Context) sdk.KVStore {
	return ctx.KVStore(k.storeKey)
}

func (k Keeper) jobsStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(k.Store(ctx), types.KeyPrefix("jobs"))
}

func (k Keeper) AddNewJob(ctx sdk.Context, job *types.Job) error {
	if k.JobIDExists(ctx, job.GetID()) {
		return types.ErrJobWithIDAlreadyExists.Wrap(job.GetID())
	}

	return k.saveJob(ctx, job)
}

func (k Keeper) saveJob(ctx sdk.Context, job *types.Job) error {
	if job.GetOwner().Empty() {
		return types.ErrInvalid.Wrap("owner can't be empty when adding a new job")
	}

	if err := job.ValidateBasic(); err != nil {
		return err
	}

	router := job.GetRouting()

	chain, ok := k.Chains[router.GetChainType()]
	if !ok {
		supportedChains := slice.FromMapKeys(k.Chains)
		return types.ErrInvalid.Wrapf("chain type %s is not supported. supported chains: %s", router.GetChainType(), supportedChains)
	}

	// unmarshaling now to test if the payload is correct
	err := chain.VerifyJob(
		ctx,
		job.GetDefinition(),
		job.GetPayload(),
		router.GetChainReferenceID(),
	)
	if err != nil {
		return whoops.Wrap(err, types.ErrInvalid)
	}

	return keeperutil.Save(k.jobsStore(ctx), k.cdc, []byte(job.GetID()), job)
}

func (k Keeper) JobIDExists(ctx sdk.Context, jobID string) bool {
	return k.jobsStore(ctx).Has([]byte(jobID))
}

func (k Keeper) getJob(ctx sdk.Context, jobID string) (*types.Job, error) {
	job, err := keeperutil.Load[*types.Job](k.jobsStore(ctx), k.cdc, []byte(jobID))
	if errors.Is(err, keeperutil.ErrNotFound) {
		return nil, types.ErrJobNotFound.Wrap("job id: " + jobID)
	}

	return job, nil
}

func (k Keeper) ScheduleNow(ctx sdk.Context, jobID string, in []byte) error {
	job, err := k.getJob(ctx, jobID)
	if err != nil {
		return err
	}

	router := job.GetRouting()

	chain := k.Chains[router.GetChainType()]

	payload := job.GetPayload()

	if len(in) > 0 && !job.GetIsPayloadModifiable() {
		return types.ErrCannotModifyJobPayload.Wrapf("jobID: %s", jobID)
	}

	if job.GetIsPayloadModifiable() && in != nil {
		payload = in
	}

	return chain.ExecuteJob(ctx, job.GetDefinition(), payload, router.GetChainReferenceID())
}
