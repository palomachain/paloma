package keeper

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/VolumeFi/whoops"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	xchain "github.com/palomachain/paloma/internal/x-chain"
	keeperutil "github.com/palomachain/paloma/util/keeper"
	"github.com/palomachain/paloma/util/liblog"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/scheduler/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	memKey     storetypes.StoreKey
	paramstore paramtypes.Subspace

	account   types.AccountKeeper
	EvmKeeper types.EvmKeeper

	ider keeperutil.IDGenerator

	Chains map[xchain.Type]xchain.Bridge
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	account types.AccountKeeper,
	evmKeeper types.EvmKeeper,
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
		account:    account,
		EvmKeeper:  evmKeeper,
		Chains:     cm,
	}

	k.ider = keeperutil.NewIDGenerator(k, nil)

	return k
}

func (k Keeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
	return k.account.GetAccount(ctx, addr)
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) ModuleName() string {
	return types.ModuleName
}

func (k Keeper) PreJobExecution(ctx sdk.Context, job *types.Job) error {
	return k.EvmKeeper.PreJobExecution(ctx, job)
}

// store returns default store for this keeper!
func (k Keeper) Store(ctx context.Context) sdk.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.KVStore(k.storeKey)
}

func (k Keeper) jobsStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(k.Store(ctx), types.KeyPrefix("jobs"))
}

func (k Keeper) AddNewJob(ctx sdk.Context, job *types.Job) error {
	if k.JobIDExists(ctx, job.GetID()) {
		return types.ErrJobWithIDAlreadyExists.Wrap(job.GetID())
	}

	err := k.saveJob(ctx, job)
	if err != nil {
		return err
	}

	keeperutil.EmitEvent(k, ctx, "JobAdded",
		sdk.NewAttribute("id", job.GetID()),
	)

	return nil
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

func (k Keeper) GetJob(ctx sdk.Context, jobID string) (*types.Job, error) {
	job, err := keeperutil.Load[*types.Job](k.jobsStore(ctx), k.cdc, []byte(jobID))
	if errors.Is(err, keeperutil.ErrNotFound) {
		return nil, types.ErrJobNotFound.Wrap("job id: " + jobID)
	}

	return job, nil
}

func (k Keeper) ExecuteJob(ctx sdk.Context, jobID string, payload []byte, senderAddress sdk.AccAddress, contractAddr sdk.AccAddress) (uint64, error) {
	job, err := k.GetJob(ctx, jobID)
	if err != nil {
		liblog.FromSDKLogger(k.Logger(ctx)).WithError(err).WithFields("job-id", jobID).Error("Job not found.")
		return 0, err
	}

	// Hook to trigger a valset update attempt
	err = k.PreJobExecution(ctx, job)
	if err != nil {
		// If we have an error here, don't exit.  Go ahead and schedule the job.
		// Paloma will try to push the valset update again with the next job.
		liblog.FromSDKLogger(k.Logger(ctx)).WithError(err).WithFields("job-id", jobID).Error("failed to run PreJobExecution hook")
	}

	return k.ScheduleNow(ctx, jobID, payload, senderAddress, contractAddr)
}

func (k Keeper) ScheduleNow(ctx sdk.Context, jobID string, in []byte, senderAddress sdk.AccAddress, contractAddress sdk.AccAddress) (uint64, error) {
	job, err := k.GetJob(ctx, jobID)
	if err != nil {
		k.Logger(ctx).Error("couldn't schedule a job", "job_id", jobID, "err", err)
		return 0, err
	}

	router := job.GetRouting()

	chain := k.Chains[router.GetChainType()]

	payload := job.GetPayload()

	k.Logger(ctx).Info(
		"scheduling a job",
		"job_id", jobID,
		"chain_type", router.GetChainType(),
		"chain_reference_id", router.GetChainReferenceID(),
	)

	if len(in) > 0 && !job.GetIsPayloadModifiable() {
		k.Logger(ctx).Error(
			"couldn't modify a job's payload as the payload is not modifiable for the job",
			"job_id", jobID,
			"err", err,
		)
		return 0, types.ErrCannotModifyJobPayload.Wrapf("jobID: %s", jobID)
	}

	if job.GetIsPayloadModifiable() && in != nil {
		payload = in
	}

	jcfg := &xchain.JobConfiguration{
		Definition:      job.GetDefinition(),
		Payload:         payload,
		SenderAddress:   senderAddress,
		ContractAddress: contractAddress,
		RefID:           router.GetChainReferenceID(),
		Requirements: xchain.JobRequirements{
			EnforceMEVRelay: job.EnforceMEVRelay,
		},
	}

	msgID, err := chain.ExecuteJob(ctx, jcfg)
	if err != nil {
		liblog.FromSDKLogger(k.Logger(ctx)).WithError(err).WithFields(
			"job_id", jobID,
			"payload", payload,
			"chain_type", router.GetChainType(),
			"chain_reference_id", router.GetChainReferenceID(),
			"enforce_mev_relay", job.EnforceMEVRelay,
		).Error("couldn't execute a job")
		return 0, err
	}

	keeperutil.EmitEvent(k, ctx, "JobScheduler",
		sdk.NewAttribute("job-id", jobID),
		sdk.NewAttribute("msg-id", strconv.FormatUint(msgID, 10)),
		sdk.NewAttribute("chainType", router.GetChainType()),
		sdk.NewAttribute("chainReferenceID", router.GetChainReferenceID()),
	)
	return msgID, nil
}
