package keeper

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	consensusmocks "github.com/palomachain/paloma/x/consensus/keeper/consensus/mocks"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
	evmmocks "github.com/palomachain/paloma/x/evm/types/mocks"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("attest submit logic call", func() {
	var k *Keeper
	var ctx sdk.Context
	var q *consensusmocks.Queuer
	var msg *consensustypes.QueuedSignedMessage
	var consensuskeeper *evmmocks.ConsensusKeeper
	var evidence []*consensustypes.Evidence
	var retries uint32

	testChain := &types.AddChainProposal{
		ChainReferenceID:  "eth-main",
		Title:             "Test Title",
		Description:       "Test description",
		BlockHeight:       uint64(123),
		BlockHashAtHeight: "0x1234",
	}

	BeforeEach(func() {
		var ms mockedServices
		k, ms, ctx = NewEvmKeeper(GinkgoT())
		consensuskeeper = ms.ConsensusKeeper
		q = consensusmocks.NewQueuer(GinkgoT())

		snapshot := createSnapshot(testChain)
		ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(snapshot, nil)

		q.On("ChainInfo").Return("", "eth-main")
		q.On("Remove", mock.Anything, uint64(123)).Return(nil)
		ms.SkywayKeeper.On("GetLastObservedSkywayNonce", mock.Anything, mock.Anything).
			Return(uint64(100), nil).Maybe()

		err := setupTestChainSupport(ctx, consensuskeeper, testChain, k)
		Expect(err).To(BeNil())
	})

	JustBeforeEach(func() {
		consensusMsg, err := codectypes.NewAnyWithValue(&types.Message{
			Action: &types.Message_SubmitLogicCall{
				SubmitLogicCall: &types.SubmitLogicCall{
					Retries: retries,
				},
			},
		})
		Expect(err).To(BeNil())

		msg = &consensustypes.QueuedSignedMessage{
			Id:       123,
			Msg:      consensusMsg,
			Evidence: evidence,
		}
	})

	Context("attesting with proof error", func() {
		BeforeEach(func() {
			proof, _ := codectypes.NewAnyWithValue(
				&types.SmartContractExecutionErrorProof{
					ErrorMessage: "an error",
				})
			evidence = []*consensustypes.Evidence{{
				ValAddress: sdk.ValAddress("addr1"),
				Proof:      proof,
			}, {
				ValAddress: sdk.ValAddress("addr2"),
				Proof:      proof,
			}}
		})

		JustBeforeEach(func() {
			Expect(k.attestRouter(ctx, q, msg)).To(Succeed())
		})

		Context("attesting with 0 retries", func() {
			BeforeEach(func() {
				retries = 0
				consensuskeeper.On("PutMessageInQueue",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(uint64(10), nil).Once()
			})

			It("should retry the deployment", func() {
				// Should be called once on setup and again on retry
				consensuskeeper.AssertNumberOfCalls(GinkgoT(), "PutMessageInQueue", 2)
			})

			It("should keep the deployment", func() {
				val, _ := k.getSmartContractDeploymentByContractID(ctx, 1,
					testChain.GetChainReferenceID())
				Expect(val).ToNot(BeNil())
			})

			It("should increase retries on the smart contract deployment", func() {
				cm, _ := msg.ConsensusMsg(k.cdc)
				action := cm.(*types.Message).Action.(*types.Message_SubmitLogicCall)
				Expect(action.SubmitLogicCall.Retries).To(BeNumerically("==", 1))
			})
		})

		Context("attesting after retry limit", func() {
			BeforeEach(func() {
				retries = 2
			})

			Context("with regular SLC", func() {
				It("should not put message back into the queue", func() {
					// Should be called only once on setup
					consensuskeeper.AssertNumberOfCalls(GinkgoT(), "PutMessageInQueue", 1)
				})
			})

			Context("with ERC20 transfer", func() {
				BeforeEach(func() {
					k.deploymentCache.Add(ctx, testChain.GetChainReferenceID(), 1, 123)
					dep, _ := k.getSmartContractDeploymentByContractID(ctx, 1, testChain.GetChainReferenceID())
					dep.Erc20Transfers = append(dep.Erc20Transfers, types.SmartContractDeployment_ERC20Transfer{
						Denom:  "test-denom",
						Erc20:  "test-denom",
						MsgID:  123,
						Status: types.SmartContractDeployment_ERC20Transfer_FAIL,
					})
					k.updateSmartContractDeployment(ctx, 1, testChain.ChainReferenceID, dep)
					consensuskeeper.On("PutMessageInQueue",
						mock.Anything,
						mock.Anything,
						mock.Anything,
						mock.Anything,
					).Return(uint64(10), nil).Once()
				})

				It("should put message back into the queue", func() {
					// Should be called only once on setup
					consensuskeeper.AssertNumberOfCalls(GinkgoT(), "PutMessageInQueue", 2)
				})
			})
		})
	})
})
