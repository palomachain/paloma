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

		chains map[xchain.Type]xchain.Bridge
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
		chains:     cm,
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

func (k Keeper) addNewJob(ctx sdk.Context, job *types.Job) error {
	if k.JobIDExists(ctx, job.GetID()) {
		return types.ErrJobWithIDAlreadyExists.Wrap(job.GetID())
	}

	return k.saveJob(ctx, job)
}

func (k Keeper) saveJob(ctx sdk.Context, job *types.Job) error {
	if job.GetOwner().Empty() {
		return types.ErrInvalid.Wrap("owner can't be empty when adding a new job")
	}

	router := job.GetRouting()

	chain := k.chains[router.GetChainType()]

	// unmarshaling now to test if the payload is correct
	_, err := chain.UnmarshalJob(job.GetDefinition(), job.GetPayload(), router.GetChainReferenceID())
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

// func (k Keeper) ScheduleNow(ctx sdk.Context, jobID string, payloadIn []byte) error {
// 	job := k.getJob(ctx, jobID)

// 	router := job.GetRouting()

// 	chain := k.chains[router.GetChainType()]

// 	payload := job.GetPayload()

// 	if job.GetIsPayloadModifiable() {
// 		payload = payloadIn
// 	}

// 	jobInfo, err = chain.UnmarshalJob(job.GetDefinition(), payload)
// 	if err != nil {
// 		return err
// 	}

// 	jobInfo, err := chain.UnmarshalJob(job.GetDefinition())
// 	if err != nil {
// 		return err
// 	}

// 	return k.Consensus.PutMessageInQueue(ctx, jobInfo.Queue, jobInfo.Msg, &consensus.PutOptions{
// 		RequireSignatures: true,
// 	})
// }
