package keeper_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestScheduler(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Scheduler")
}

var _ = Describe("jailing validators with missing external chain infos", func() {

	// var k *keeper.Keeper
	// var ctx sdk.Context
	// var a app.TestApp

	// BeforeEach(func() {
	// 	t := GinkgoT()
	// 	a = app.NewTestApp(t, false)
	// 	ctx = a.NewContext(false, tmproto.Header{
	// 		Height: 5,
	// 	})
	// 	k = &a.PalomaKeeper
	// })
})
