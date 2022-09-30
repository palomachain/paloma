package xchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/vizualni/whoops"
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
	XChainReferenceIDs(sdk.Context) []ReferenceID
}

type CobraTXJobAdder interface {
	Args() cobra.Command
	Flags() any
	Parse(*cobra.Command, []string) (any, error)
}

type Jobber interface {
	VerifyJob(ctx sdk.Context, definition []byte, payload []byte, refID ReferenceID) (err error)
	ExecuteJob(ctx sdk.Context, definition []byte, payload []byte, refID ReferenceID) (err error)
}

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
