package keeper

import (
	"encoding/hex"
	"math/big"
	"os"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethcoretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	consensusmocks "github.com/palomachain/paloma/v2/x/consensus/keeper/consensus/mocks"
	consensustypes "github.com/palomachain/paloma/v2/x/consensus/types"
	"github.com/palomachain/paloma/v2/x/evm/types"
	evmmocks "github.com/palomachain/paloma/v2/x/evm/types/mocks"
	metrixtypes "github.com/palomachain/paloma/v2/x/metrix/types"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("attest upload user smart contract", func() {
	var (
		k               *Keeper
		ms              mockedServices
		ctx             sdk.Context
		q               *consensusmocks.Queuer
		msg             *consensustypes.QueuedSignedMessage
		consensuskeeper *evmmocks.ConsensusKeeper
		evidence        []*consensustypes.Evidence
		retries         uint32
	)

	valAddr := "cosmosvaloper1pzf9apnk8yw7pjw3v9vtmxvn6guhkslanh8r07"

	testChain := &types.AddChainProposal{
		ChainReferenceID:  "eth-main",
		Title:             "Test Title",
		Description:       "Test description",
		BlockHeight:       uint64(123),
		BlockHashAtHeight: "0x1234",
	}

	compassABI, _ := os.ReadFile("testdata/sample-abi.json")
	compassBytecode, _ := os.ReadFile("testdata/sample-bytecode.out")
	txData, _ := os.ReadFile("testdata/user-smart-contract-tx-data.hex")
	rawData, _ := os.ReadFile("testdata/deployed-contract-event.hex")
	deployedContractEventData, _ := hex.DecodeString(string(rawData))

	testContract := &types.UserSmartContract{
		Author:           valAddr,
		Title:            "Test Contract",
		AbiJson:          string(compassABI),
		Bytecode:         string(compassBytecode),
		ConstructorInput: "0x00",
	}

	BeforeEach(func() {
		k, ms, ctx = NewEvmKeeper(GinkgoT())
		consensuskeeper = ms.ConsensusKeeper
		q = consensusmocks.NewQueuer(GinkgoT())

		snapshot := createSnapshot(testChain)
		ms.ValsetKeeper.On("GetCurrentSnapshot", mock.Anything).Return(snapshot, nil)

		q.On("ChainInfo").Return("", testChain.ChainReferenceID)
		q.On("Remove", mock.Anything, uint64(123)).Return(nil)
		ms.SkywayKeeper.On("GetLastObservedSkywayNonce", mock.Anything, mock.Anything).
			Return(uint64(100), nil).Maybe()

		contractID, err := setupTestChainSupport(ctx, consensuskeeper, ms.MetrixKeeper, ms.TreasuryKeeper, testChain, k)
		Expect(err).To(BeNil())

		deployment, _ := k.getSmartContractDeploymentByContractID(ctx, contractID, testChain.ChainReferenceID)
		deployment.Status = types.SmartContractDeployment_WAITING_FOR_ERC20_OWNERSHIP_TRANSFER

		err = k.updateSmartContractDeployment(ctx, contractID, testChain.ChainReferenceID, deployment)
		Expect(err).To(BeNil())

		err = k.SetSmartContractAsActive(ctx, contractID, testChain.ChainReferenceID)
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

		// We need more calls to these two methods here because of the user
		// smart contract upload
		ms.TreasuryKeeper.On("GetRelayerFeesByChainReferenceID", mock.Anything, mock.Anything).Return(getFees(3), nil).Once()
		ms.MetrixKeeper.On("Validators", mock.Anything, mock.Anything).Return(&metrixtypes.QueryValidatorsResponse{
			ValMetrics: getMetrics(3),
		}, nil).Once()

		// Create the deployment
		_, err = k.CreateUserSmartContractDeployment(ctx, valAddr, 1, testChain.ChainReferenceID)
		Expect(err).To(BeNil())
	})

	JustBeforeEach(func() {
		bytecode, _ := hex.DecodeString("0x00")
		senderAddr, _ := sdk.ValAddressFromBech32(valAddr)

		consensusMsg, err := codectypes.NewAnyWithValue(&types.Message{
			Action: &types.Message_UploadUserSmartContract{
				UploadUserSmartContract: &types.UploadUserSmartContract{
					Bytecode:      bytecode,
					Id:            1,
					SenderAddress: senderAddr,
					BlockHeight:   ctx.BlockHeight(),
					Retries:       retries,
					Fees: &types.Fees{
						RelayerFee:   1,
						CommunityFee: 2,
						SecurityFee:  3,
					},
				},
			},
		})
		Expect(err).To(BeNil())

		sig := make([]byte, 100)
		msg = &consensustypes.QueuedSignedMessage{
			Id:       123,
			Msg:      consensusMsg,
			Evidence: evidence,
			SignData: []*consensustypes.SignData{{
				ExternalAccountAddress: "addr1",
				Signature:              sig,
			}, {
				ExternalAccountAddress: "addr2",
				Signature:              sig,
			}},
		}
	})

	Context("attesting with success proof", func() {
		BeforeEach(func() {
			tx := ethcoretypes.NewTx(&ethcoretypes.DynamicFeeTx{
				ChainID: big.NewInt(1),
				Data:    common.FromHex(string(txData)),
			})

			signer := ethcoretypes.LatestSignerForChainID(big.NewInt(1))
			privkey, err := crypto.GenerateKey()
			Expect(err).To(BeNil())

			signature, err := crypto.Sign(tx.Hash().Bytes(), privkey)
			Expect(err).To(BeNil())

			signedTX, err := tx.WithSignature(signer, signature)
			Expect(err).To(BeNil())

			serializedTX, err := signedTX.MarshalBinary()
			Expect(err).To(BeNil())

			receipt := ethcoretypes.Receipt{
				Status: ethcoretypes.ReceiptStatusSuccessful,
				Logs: []*ethcoretypes.Log{
					{
						Topics: []common.Hash{contractDeployedEvent},
						Data:   deployedContractEventData,
					},
				},
			}
			serializedReceipt, err := receipt.MarshalBinary()
			Expect(err).To(BeNil())

			proof, _ := codectypes.NewAnyWithValue(
				&types.TxExecutedProof{
					SerializedTX:      serializedTX,
					SerializedReceipt: serializedReceipt,
				})
			evidence = []*consensustypes.Evidence{{
				ValAddress: sdk.ValAddress("validator-1"),
				Proof:      proof,
			}, {
				ValAddress: sdk.ValAddress("validator-2"),
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
					Status:           types.UserSmartContract_Deployment_ACTIVE,
					Address:          "0x5eeA9CdF6de50497Df2c5AF93B70A02D616454a0",
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
				ValAddress: sdk.ValAddress("validator-1"),
				Proof:      proof,
			}, {
				ValAddress: sdk.ValAddress("validator-2"),
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

				// We need to setup additional calls for the retry
				ms.TreasuryKeeper.On("GetRelayerFeesByChainReferenceID", mock.Anything, mock.Anything).Return(getFees(3), nil).Once()
				ms.MetrixKeeper.On("Validators", mock.Anything, mock.Anything).Return(&metrixtypes.QueryValidatorsResponse{
					ValMetrics: getMetrics(3),
				}, nil).Once()
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
						Status:           types.UserSmartContract_Deployment_IN_FLIGHT,
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
						Status:           types.UserSmartContract_Deployment_ERROR,
						Address:          "",
					},
				))
			})
		})
	})
})
