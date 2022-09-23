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

//go:generate mockery --name=Bridge
type Bridge interface {
	Info
	JobMarshaller
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

type JobInfo struct {
	Queue      string
	Definition any
}

type JobMarshaller interface {
	// TODO: rename to job verifier
	UnmarshalJob(definition []byte, payload []byte, refID ReferenceID) (jobWithPayload JobInfo, err error)
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
