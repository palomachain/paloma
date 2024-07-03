package keeper

import (
	"math/big"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	consensusmocks "github.com/palomachain/paloma/x/consensus/keeper/consensus/mocks"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
	evmmocks "github.com/palomachain/paloma/x/evm/types/mocks"
	"github.com/stretchr/testify/mock"
)

func setupTestChainSupport(
	ctx sdk.Context,
	consensuskeeper *evmmocks.ConsensusKeeper,
	chain *types.AddChainProposal,
	k *Keeper,
) error {
	consensuskeeper.On("PutMessageInQueue",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(uint64(10), nil).Once()

	err := k.AddSupportForNewChain(
		ctx,
		chain.GetChainReferenceID(),
		chain.GetChainID(),
		chain.GetBlockHeight(),
		chain.GetBlockHashAtHeight(),
		big.NewInt(55),
	)
	if err != nil {
		return err
	}

	sc, err := k.SaveNewSmartContract(ctx, contractAbi, common.FromHex(contractBytecodeStr))
	if err != nil {
		return err
	}

	return k.SetAsCompassContract(ctx, sc)
}

var _ = Describe("attest upload smart contract", func() {
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
			Action: &types.Message_UploadSmartContract{
				UploadSmartContract: &types.UploadSmartContract{
					Id:      1,
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
				action := cm.(*types.Message).Action.(*types.Message_UploadSmartContract)
				Expect(action.UploadSmartContract.Retries).To(BeNumerically("==", 1))
			})
		})

		Context("attesting after retry limit", func() {
			BeforeEach(func() {
				retries = 2
			})

			It("should not put message back into the queue", func() {
				// Should be called only once on setup
				consensuskeeper.AssertNumberOfCalls(GinkgoT(), "PutMessageInQueue", 1)
			})

			It("should remove the smart contract deployment", func() {
				val, _ := k.getSmartContractDeploymentByContractID(ctx, 1,
					testChain.GetChainReferenceID())
				Expect(val).To(BeNil())
			})
		})
	})
})
