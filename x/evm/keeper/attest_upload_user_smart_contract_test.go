package keeper

import (
	"encoding/hex"

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

var _ = Describe("attest upload user smart contract", func() {
	var (
		k               *Keeper
		ctx             sdk.Context
		q               *consensusmocks.Queuer
		msg             *consensustypes.QueuedSignedMessage
		consensuskeeper *evmmocks.ConsensusKeeper
		evidence        []*consensustypes.Evidence
		retries         uint32
	)
	valAddr := "palomavaloper1tsu8nthuspe4zlkejtj3v27rtq8qz7q6983zt2"

	testChain := &types.AddChainProposal{
		ChainReferenceID:  "eth-main",
		Title:             "Test Title",
		Description:       "Test description",
		BlockHeight:       uint64(123),
		BlockHashAtHeight: "0x1234",
	}

	testContract := &types.UserSmartContract{
		ValAddress:       valAddr,
		Title:            "Test Contract",
		AbiJson:          "[{}]",
		Bytecode:         "0x01",
		ConstructorInput: "0x00",
	}

	BeforeEach(func() {
		var ms mockedServices
		k, ms, ctx = NewEvmKeeper(GinkgoT())
		consensuskeeper = ms.ConsensusKeeper
		q = consensusmocks.NewQueuer(GinkgoT())

		setupValsetKeeper(ms, testChain)

		q.On("ChainInfo").Return("", testChain.ChainReferenceID)
		q.On("Remove", mock.Anything, uint64(123)).Return(nil)
		ms.GravityKeeper.On("GetLastObservedGravityNonce", mock.Anything, mock.Anything).
			Return(uint64(100), nil).Maybe()

		err := setupTestChainSupport(ctx, consensuskeeper, testChain, k)
		Expect(err).To(BeNil())

		// Upload the contract
		_, err = k.SaveUserSmartContract(ctx, valAddr, testContract)
		Expect(err).To(BeNil())

		consensuskeeper.On("PutMessageInQueue",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(uint64(10), nil).Once()
		// Create the deployment
		_, err = k.CreateUserSmartContractDeployment(ctx, valAddr, 1, testChain.ChainReferenceID)
		Expect(err).To(BeNil())
	})

	JustBeforeEach(func() {
		consensusMsg, err := codectypes.NewAnyWithValue(&types.Message{
			Action: &types.Message_UploadUserSmartContract{
				UploadUserSmartContract: &types.UploadUserSmartContract{
					Id:         1,
					ValAddress: valAddr,
					Retries:    retries,
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

	Context("attesting with success proof", func() {
		BeforeEach(func() {
			serializedTx, _ := hex.DecodeString("02f87201108405f5e100850b68a0aa00825208941f9c2e67dbbe4c457a5e2be0bc31e67ce5953a2d87470de4df82000080c001a0e05de0771f8d577ec5aa440612c0e8f560d732d5162db0187cfaf56ac50c3716a0147565f4b0924a5adda25f55330c385448e0507d1219d4dac0950e2872682124")

			proof, _ := codectypes.NewAnyWithValue(
				&types.TxExecutedProof{
					SerializedTX: serializedTx,
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

		It("should set the deployment information", func() {
			contracts, err := k.UserSmartContracts(ctx, valAddr)
			Expect(err).To(BeNil())
			Expect(contracts[0].Deployments).To(ConsistOf(
				&types.UserSmartContract_Deployment{
					ChainReferenceId: testChain.ChainReferenceID,
					Status:           types.DeploymentStatus_ACTIVE,
					Address:          "0x069A36eC9F812D599B558fC53b81Cc039d656153",
				},
			))
		})
	})

	Context("attesting with error proof", func() {
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
				// Should be called twice on setup and again on retry
				consensuskeeper.AssertNumberOfCalls(GinkgoT(), "PutMessageInQueue", 3)
			})

			It("should keep the deployment status", func() {
				contracts, err := k.UserSmartContracts(ctx, valAddr)
				Expect(err).To(BeNil())
				Expect(contracts[0].Deployments).To(ConsistOf(
					&types.UserSmartContract_Deployment{
						ChainReferenceId: testChain.ChainReferenceID,
						Status:           types.DeploymentStatus_IN_FLIGHT,
						Address:          "",
					},
				))
			})

			It("should increase retries on the smart contract deployment", func() {
				cm, _ := msg.ConsensusMsg(k.cdc)
				action := cm.(*types.Message).Action.(*types.Message_UploadUserSmartContract)
				Expect(action.UploadUserSmartContract.Retries).To(BeNumerically("==", 1))
			})
		})

		Context("attesting after retry limit", func() {
			BeforeEach(func() {
				retries = 2
			})

			It("should not put message back into the queue", func() {
				// Should be called only twice on setup
				consensuskeeper.AssertNumberOfCalls(GinkgoT(), "PutMessageInQueue", 2)
			})

			It("should set the deployment status to error", func() {
				contracts, err := k.UserSmartContracts(ctx, valAddr)
				Expect(err).To(BeNil())
				Expect(contracts[0].Deployments).To(ConsistOf(
					&types.UserSmartContract_Deployment{
						ChainReferenceId: testChain.ChainReferenceID,
						Status:           types.DeploymentStatus_ERROR,
						Address:          "",
					},
				))
			})
		})
	})
})
