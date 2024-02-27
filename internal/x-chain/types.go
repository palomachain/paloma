package xchain

import (
	"context"

	"github.com/VolumeFi/whoops"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

type (
	Type        = string
	ReferenceID = string
)

type RequiredInterfaceToSupport interface {
	Bridge
}

//go:generate mockery --name=Bridge
type Bridge interface {
	Info
	Jobber
	WalletUpdater
}

type Info interface {
	XChainType() Type
	XChainReferenceIDs(context.Context) []ReferenceID
}

type CobraTXJobAdder interface {
	Args() cobra.Command
	Flags() any
	Parse(*cobra.Command, []string) (any, error)
}

type JobRequirements struct {
	EnforceMEVRelay bool
}

type JobConfiguration struct {
	Definition      []byte
	Payload         []byte
	SenderAddress   sdk.AccAddress
	ContractAddress sdk.AccAddress
	RefID           ReferenceID
	Requirements    JobRequirements
}

type Jobber interface {
	VerifyJob(ctx context.Context, definition []byte, payload []byte, refID ReferenceID) (err error)
	ExecuteJob(ctx context.Context, jcfg *JobConfiguration) (msgID uint64, err error)
}

//go:generate mockery --name=FundCollecter
type FundCollecter interface {
	CollectJobFundEvents(ctx sdk.Context) error
}

type WalletUpdater interface {
	// every chain that we support should implement this. Using this, we can
	// ask all chains to send wallet updates to paloma.
	// TriggerWalletUpdate(sdk.Context) error
}

// errors

const (
	ErrIncorrectJobType = whoops.Errorf("incorrect job type: got %T, expected %T")
)
