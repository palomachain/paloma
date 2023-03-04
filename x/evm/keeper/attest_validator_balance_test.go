package keeper

import (
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/palomachain/paloma/util/slice"
	"github.com/palomachain/paloma/x/evm/types"
	evmmocks "github.com/palomachain/paloma/x/evm/types/mocks"
	"github.com/stretchr/testify/mock"
)

var _ = g.Describe("attest validator balance", func() {
	var k Keeper
	var ctx sdk.Context
	var v *evmmocks.ValsetKeeper

	var subject func() error

	g.BeforeEach(func() {
		t := g.GinkgoT()
		kpr, ms, _ctx := NewEvmKeeper(t)
		ctx = _ctx
		k = *kpr
		v = ms.ValsetKeeper
	})

	var req *types.ValidatorBalancesAttestation
	var evidenceThatWon *types.ValidatorBalancesAttestationRes
	var minBalance *big.Int

	g.JustBeforeEach(func() {
		subject = func() error {
			return k.processValidatorBalanceProof(
				ctx,
				req,
				evidenceThatWon,
				"chain-id",
				minBalance,
			)
		}
	})

	g.Context("with various balances", func() {
		g.BeforeEach(func() {
			minBalance = big.NewInt(50)
			validatorBalances := []string{"51", "1000", "50", "49", "0"}
			req = &types.ValidatorBalancesAttestation{
				FromBlockTime: time.Unix(999, 0),
				HexAddresses: slice.Map(validatorBalances, func(b string) string {
					return common.BytesToAddress([]byte("addr_" + b)).Hex()
				}),
				ValAddresses: slice.Map(validatorBalances, func(b string) sdk.ValAddress {
					return sdk.ValAddress("addr_" + b)
				}),
			}
			evidenceThatWon = &types.ValidatorBalancesAttestationRes{
				BlockHeight: 123,
				Balances:    validatorBalances,
			}

			for _, val := range req.ValAddresses {
				v.On("IsJailed", mock.Anything, val).Return(false).Maybe()
			}

			v.On("Jail", mock.Anything, req.ValAddresses[3], mock.Anything).Return(nil)
			v.On("Jail", mock.Anything, req.ValAddresses[4], mock.Anything).Return(nil)

			for i := range validatorBalances {
				n, _ := new(big.Int).SetString(validatorBalances[i], 10)
				v.On(
					"SetValidatorBalance",
					mock.Anything,
					req.ValAddresses[i],
					"evm",
					"chain-id",
					req.HexAddresses[i],
					n,
				).Return(nil)
			}
		})
		g.It("updates the balances and jail those that have below min threshold", func() {
			err := subject()
			Expect(err).To(BeNil())
		})
	})
})
