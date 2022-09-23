package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/palomachain/paloma/app"
	"github.com/palomachain/paloma/x/scheduler/keeper"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestScheduler(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Scheduler")
}

var _ = Describe("jailing validators with missing external chain infos", func() {

	var k *keeper.Keeper
	var ctx sdk.Context
	var a app.TestApp

	var vals []stakingtypes.Validator

	BeforeEach(func() {
		t := GinkgoT()
		a = app.NewTestApp(t, false)
		ctx = a.NewContext(false, tmproto.Header{
			Height: 5,
		})
		k = &a.PalomaKeeper
	})
})
