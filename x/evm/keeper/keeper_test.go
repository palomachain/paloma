package keeper_test

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"
	"testing"

	"github.com/VolumeFi/whoops"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/palomachain/paloma/app"
	testutil "github.com/palomachain/paloma/testutil"
	"github.com/palomachain/paloma/testutil/rand"
	"github.com/palomachain/paloma/testutil/sample"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/keeper"
	"github.com/palomachain/paloma/x/evm/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	contractAbi         = string(whoops.Must(ioutil.ReadFile("testdata/sample-abi.json")))
	contractBytecodeStr = string(whoops.Must(ioutil.ReadFile("testdata/sample-bytecode.out")))
)

func genValidators(numValidators, totalConsPower int) []stakingtypes.Validator {
	return testutil.GenValidators(numValidators, totalConsPower)
}

func TestEndToEndForEvmArbitraryCall(t *testing.T) {
	chainType, chainReferenceID := consensustypes.ChainTypeEVM, "eth-main"
	a := app.NewTestApp(t, false)
	ctx := a.NewContext(false, tmproto.Header{
		Height: 5,
	})

	newChain := &types.AddChainProposal{
		ChainReferenceID:  "eth-main",
		Title:             "bla",
		Description:       "bla",
		BlockHeight:       uint64(123),
		BlockHashAtHeight: "0x1234",
	}

	err := a.EvmKeeper.AddSupportForNewChain(
		ctx,
		newChain.GetChainReferenceID(),
		newChain.GetChainID(),
		newChain.GetBlockHeight(),
		newChain.GetBlockHashAtHeight(),
		big.NewInt(55),
	)
	require.NoError(t, err)

	err = a.EvmKeeper.ActivateChainReferenceID(ctx, newChain.ChainReferenceID, &types.SmartContract{Id: 123}, "addr", []byte("abc"))
	require.NoError(t, err)

	validators := genValidators(25, 25000)
	for _, val := range validators {
		a.StakingKeeper.SetValidator(ctx, val)
	}

	smartContractAddr := common.BytesToAddress(rand.Bytes(5))
	err = a.EvmKeeper.AddSmartContractExecutionToConsensus(
		ctx,
		chainReferenceID,
		"",
		&types.SubmitLogicCall{
			Payload: func() []byte {
				evm := whoops.Must(abi.JSON(strings.NewReader(sample.SimpleABI)))
				return whoops.Must(evm.Pack("store", big.NewInt(1337)))
			}(),
			HexContractAddress: smartContractAddr.Hex(),
			Abi:                []byte(sample.SimpleABI),
			Deadline:           1337,
		},
	)

	require.NoError(t, err)

	private, err := crypto.GenerateKey()
	require.NoError(t, err)

	accAddr := crypto.PubkeyToAddress(private.PublicKey)
	err = a.ValsetKeeper.AddExternalChainInfo(ctx, validators[0].GetOperator(), []*valsettypes.ExternalChainInfo{
		{
			ChainType:        chainType,
			ChainReferenceID: chainReferenceID,
			Address:          accAddr.Hex(),
			Pubkey:           accAddr[:],
		},
	})

	require.NoError(t, err)
	queue := consensustypes.Queue(keeper.ConsensusTurnstoneMessage, chainType, chainReferenceID)
	msgs, err := a.ConsensusKeeper.GetMessagesForSigning(ctx, queue, validators[0].GetOperator())

	for _, msg := range msgs {
		sigbz, err := crypto.Sign(
			crypto.Keccak256(
				[]byte(keeper.SignaturePrefix),
				msg.GetBytesToSign(),
			),
			private,
		)
		require.NoError(t, err)
		err = a.ConsensusKeeper.AddMessageSignature(
			ctx,
			validators[0].GetOperator(),
			[]*consensustypes.ConsensusMessageSignature{
				{
					Id:              msg.GetId(),
					QueueTypeName:   queue,
					Signature:       sigbz,
					SignedByAddress: accAddr.Hex(),
				},
			},
		)
		require.NoError(t, err)
	}

}

func TestOnSnapshotBuilt(t *testing.T) {
	a := app.NewTestApp(t, false)
	ctx := a.NewContext(false, tmproto.Header{
		Height: 5,
	})

	newChain := &types.AddChainProposal{
		ChainReferenceID:  "bob",
		Title:             "bla",
		Description:       "bla",
		BlockHeight:       uint64(123),
		BlockHashAtHeight: "0x1234",
	}
	err := a.EvmKeeper.AddSupportForNewChain(
		ctx,
		newChain.GetChainReferenceID(),
		newChain.GetChainID(),
		newChain.GetBlockHeight(),
		newChain.GetBlockHashAtHeight(),
		big.NewInt(55),
	)
	require.NoError(t, err)
	err = a.EvmKeeper.ActivateChainReferenceID(
		ctx,
		newChain.ChainReferenceID,
		&types.SmartContract{
			Id: 123,
		},
		"addr",
		[]byte("abc"),
	)
	require.NoError(t, err)

	validators := genValidators(25, 25000)
	for _, val := range validators {
		a.StakingKeeper.SetValidator(ctx, val)
		err = a.ValsetKeeper.AddExternalChainInfo(ctx, val.GetOperator(), []*valsettypes.ExternalChainInfo{
			{
				ChainType:        "evm",
				ChainReferenceID: "bob",
				Address:          rand.ETHAddress().Hex(),
				Pubkey:           []byte("pk" + rand.ETHAddress().Hex()),
			},
		})
		require.NoError(t, err)
	}

	queue := fmt.Sprintf("evm/%s/%s", newChain.GetChainReferenceID(), keeper.ConsensusTurnstoneMessage)

	msgs, err := a.ConsensusKeeper.GetMessagesFromQueue(ctx, queue, 1)
	require.NoError(t, err)
	require.Empty(t, msgs)

	_, err = a.ValsetKeeper.TriggerSnapshotBuild(ctx)
	require.NoError(t, err)

	msgs, err = a.ConsensusKeeper.GetMessagesFromQueue(ctx, queue, 1)
	require.NoError(t, err)
	require.Len(t, msgs, 1)

}

func TestAddingSupportForNewChain(t *testing.T) {
	a := app.NewTestApp(t, false)
	ctx := a.NewContext(false, tmproto.Header{
		Height: 5,
	})

	t.Run("with happy path there are no errors", func(t *testing.T) {
		newChain := &types.AddChainProposal{
			ChainReferenceID:  "bob",
			Title:             "bla",
			Description:       "bla",
			BlockHeight:       uint64(123),
			BlockHashAtHeight: "0x1234",
		}
		err := a.EvmKeeper.AddSupportForNewChain(
			ctx,
			newChain.GetChainReferenceID(),
			newChain.GetChainID(),
			newChain.GetBlockHeight(),
			newChain.GetBlockHashAtHeight(),
			big.NewInt(55),
		)
		require.NoError(t, err)

		gotChainInfo, err := a.EvmKeeper.GetChainInfo(ctx, newChain.GetChainReferenceID())
		require.NoError(t, err)

		require.Equal(t, newChain.GetChainReferenceID(), gotChainInfo.GetChainReferenceID())
		require.Equal(t, newChain.GetBlockHashAtHeight(), gotChainInfo.GetReferenceBlockHash())
		require.Equal(t, newChain.GetBlockHeight(), gotChainInfo.GetReferenceBlockHeight())
		t.Run("it returns an error if we try to add a chian whose chainID already exists", func(t *testing.T) {
			newChain.ChainReferenceID = "something_new"
			err := a.EvmKeeper.AddSupportForNewChain(
				ctx,
				newChain.GetChainReferenceID(),
				newChain.GetChainID(),
				newChain.GetBlockHeight(),
				newChain.GetBlockHashAtHeight(),
				big.NewInt(55),
			)
			require.ErrorIs(t, err, keeper.ErrCannotAddSupportForChainThatExists)
		})
	})

	t.Run("when chainReferenceID already exists then it returns an error", func(t *testing.T) {
		newChain := &types.AddChainProposal{
			ChainReferenceID:  "bob",
			Title:             "bla",
			Description:       "bla",
			BlockHeight:       uint64(123),
			BlockHashAtHeight: "0x1234",
		}
		err := a.EvmKeeper.AddSupportForNewChain(
			ctx,
			newChain.GetChainReferenceID(),
			newChain.GetChainID(),
			newChain.GetBlockHeight(),
			newChain.GetBlockHashAtHeight(),

			big.NewInt(55),
		)
		require.Error(t, err)
	})

	t.Run("activiting chain", func(t *testing.T) {
		t.Run("if the chain does not exist it returns the error", func(t *testing.T) {
			err := a.EvmKeeper.ActivateChainReferenceID(ctx, "i don't exist", &types.SmartContract{}, "", []byte{})
			require.Error(t, err)
		})
		t.Run("works when chain exists", func(t *testing.T) {
			err := a.EvmKeeper.ActivateChainReferenceID(ctx, "bob", &types.SmartContract{Id: 123}, "addr", []byte("unique id"))
			require.NoError(t, err)
			gotChainInfo, err := a.EvmKeeper.GetChainInfo(ctx, "bob")
			require.NoError(t, err)

			require.Equal(t, "addr", gotChainInfo.GetSmartContractAddr())
			require.Equal(t, []byte("unique id"), gotChainInfo.GetSmartContractUniqueID())
		})
	})

	t.Run("removing chain", func(t *testing.T) {
		t.Run("if the chain does not exist it returns the error", func(t *testing.T) {
			err := a.EvmKeeper.RemoveSupportForChain(ctx, &types.RemoveChainProposal{
				ChainReferenceID: "i don't exist",
			})
			require.Error(t, err)
		})
		t.Run("works when chain exists", func(t *testing.T) {
			err := a.EvmKeeper.RemoveSupportForChain(ctx, &types.RemoveChainProposal{
				ChainReferenceID: "bob",
			})
			require.NoError(t, err)
			_, err = a.EvmKeeper.GetChainInfo(ctx, "bob")
			require.Error(t, keeper.ErrChainNotFound)
		})
	})
}

func TestWithGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "EVM keeper")
}

var _ = Describe("evm", func() {
	// smartContractAddr := common.BytesToAddress(rand.Bytes(5))
	// chainType, chainReferenceID := consensustypes.ChainTypeEVM, "eth-main"
	var a app.TestApp
	var ctx sdk.Context
	var validators []stakingtypes.Validator
	newChain := &types.AddChainProposal{
		ChainReferenceID:  "eth-main",
		Title:             "bla",
		Description:       "bla",
		BlockHeight:       uint64(123),
		BlockHashAtHeight: "0x1234",
	}
	smartContract := &types.SmartContract{
		Id:       1,
		AbiJSON:  contractAbi,
		Bytecode: common.FromHex(contractBytecodeStr),
	}
	smartContract2 := &types.SmartContract{
		Id:       2,
		AbiJSON:  contractAbi,
		Bytecode: common.FromHex(contractBytecodeStr),
	}

	BeforeEach(func() {
		a = app.NewTestApp(GinkgoT(), false)
		ctx = a.NewContext(false, tmproto.Header{
			Height: 5,
		})
	})

	Context("multiple chains and smart contracts", func() {
		Describe("trying to add support for the same chain twice", func() {
			It("returns an error", func() {
				err := a.EvmKeeper.AddSupportForNewChain(
					ctx,
					newChain.GetChainReferenceID(),
					newChain.GetChainID(),
					newChain.GetBlockHeight(),
					newChain.GetBlockHashAtHeight(),
					big.NewInt(55),
				)
				Expect(err).To(BeNil())

				err = a.EvmKeeper.AddSupportForNewChain(
					ctx,
					newChain.GetChainReferenceID(),
					newChain.GetChainID(),
					newChain.GetBlockHeight(),
					newChain.GetBlockHashAtHeight(),
					big.NewInt(55),
				)
				Expect(err).To(MatchError(keeper.ErrCannotAddSupportForChainThatExists))
			})
		})

		Describe("ensuring that there can be two chains at the same time", func() {
			chain1 := &types.AddChainProposal{
				ChainReferenceID:  "chain1",
				Title:             "bla",
				Description:       "bla",
				BlockHeight:       uint64(456),
				BlockHashAtHeight: "0x1234",
				ChainID:           1,
			}
			chain2 := &types.AddChainProposal{
				ChainReferenceID:  "chain2",
				Title:             "bla",
				Description:       "bla",
				BlockHeight:       uint64(123),
				BlockHashAtHeight: "0x5678",
				ChainID:           2,
			}
			BeforeEach(func() {
				validators = genValidators(25, 25000)
				for _, val := range validators {
					a.StakingKeeper.SetValidator(ctx, val)
				}
			})

			JustBeforeEach(func() {
				for _, val := range validators {
					private1, err := crypto.GenerateKey()
					private2, err := crypto.GenerateKey()
					Expect(err).To(BeNil())
					accAddr1 := crypto.PubkeyToAddress(private1.PublicKey)
					accAddr2 := crypto.PubkeyToAddress(private2.PublicKey)
					err = a.ValsetKeeper.AddExternalChainInfo(ctx, val.GetOperator(), []*valsettypes.ExternalChainInfo{
						{
							ChainType:        "evm",
							ChainReferenceID: chain1.ChainReferenceID,
							Address:          accAddr1.Hex(),
							Pubkey:           []byte("pub key 1" + accAddr1.Hex()),
						},
						{
							ChainType:        "evm",
							ChainReferenceID: chain2.ChainReferenceID,
							Address:          accAddr2.Hex(),
							Pubkey:           []byte("pub key 2" + accAddr2.Hex()),
						},
					})
					Expect(err).To(BeNil())
				}
				_, err := a.ValsetKeeper.TriggerSnapshotBuild(ctx)
				Expect(err).To(BeNil())
			})

			BeforeEach(func() {
				By("adding chain1 works")
				err := a.EvmKeeper.AddSupportForNewChain(
					ctx,
					chain1.GetChainReferenceID(),
					chain1.GetChainID(),
					chain1.GetBlockHeight(),
					chain1.GetBlockHashAtHeight(),
					big.NewInt(55),
				)
				Expect(err).To(BeNil())

				By("adding chain2 works")
				err = a.EvmKeeper.AddSupportForNewChain(
					ctx,
					chain2.GetChainReferenceID(),
					chain2.GetChainID(),
					chain2.GetBlockHeight(),
					chain2.GetBlockHashAtHeight(),
					big.NewInt(55),
				)
				Expect(err).To(BeNil())
			})

			Context("adding smart contract", func() {

				It("adds a new smart contract deployment", func() {
					By("simple assertion that two smart contracts share different ids", func() {
						Expect(smartContract.GetId()).NotTo(Equal(smartContract2.GetId()))
					})
					By("saving a new smart contract", func() {
						Expect(
							a.EvmKeeper.HasAnySmartContractDeployment(ctx, chain1.GetChainReferenceID()),
						).To(BeFalse())
						Expect(
							a.EvmKeeper.HasAnySmartContractDeployment(ctx, chain2.GetChainReferenceID()),
						).To(BeFalse())

						_, err := a.EvmKeeper.SaveNewSmartContract(ctx, smartContract.GetAbiJSON(), smartContract.GetBytecode())
						Expect(err).To(BeNil())

						Expect(
							a.EvmKeeper.HasAnySmartContractDeployment(ctx, chain1.GetChainReferenceID()),
						).To(BeTrue())
						Expect(
							a.EvmKeeper.HasAnySmartContractDeployment(ctx, chain2.GetChainReferenceID()),
						).To(BeTrue())
					})

					By("removing a smart deployment for chain1 - it means that it was successfuly uploaded", func() {
						a.EvmKeeper.RemoveSmartContractDeployment(ctx, smartContract.GetId(), chain1.GetChainReferenceID())
						Expect(
							a.EvmKeeper.HasAnySmartContractDeployment(ctx, chain1.GetChainReferenceID()),
						).To(BeFalse())
						Expect(
							a.EvmKeeper.HasAnySmartContractDeployment(ctx, chain2.GetChainReferenceID()),
						).To(BeTrue())

					})

					By("activating a new smart contract it removes a deployment for chain1 but it doesn't for chain2", func() {
						err := a.EvmKeeper.ActivateChainReferenceID(ctx, chain1.GetChainReferenceID(), smartContract, "addr1", []byte("id1"))
						Expect(err).To(BeNil())
						Expect(
							a.EvmKeeper.HasAnySmartContractDeployment(ctx, chain1.GetChainReferenceID()),
						).To(BeFalse())
						Expect(
							a.EvmKeeper.HasAnySmartContractDeployment(ctx, chain2.GetChainReferenceID()),
						).To(BeTrue())

						By("verify that the chain's smart contract id has been deployed", func() {
							ci, err := a.EvmKeeper.GetChainInfo(ctx, chain1.GetChainReferenceID())
							Expect(err).To(BeNil())
							Expect(ci.GetActiveSmartContractID()).To(Equal(smartContract.GetId()))
						})
					})

					By("adding a new smart contract deployment deploys it to chain1 only", func() {
						_, err := a.EvmKeeper.SaveNewSmartContract(ctx, smartContract2.GetAbiJSON(), smartContract2.GetBytecode())
						Expect(err).To(BeNil())
						Expect(
							a.EvmKeeper.HasAnySmartContractDeployment(ctx, chain1.GetChainReferenceID()),
						).To(BeTrue())
					})

					By("activating a new-new smart contract it deploys it to chain 1", func() {
						err := a.EvmKeeper.ActivateChainReferenceID(ctx, chain1.GetChainReferenceID(), smartContract2, "addr2", []byte("id2"))
						Expect(err).To(BeNil())
						Expect(
							a.EvmKeeper.HasAnySmartContractDeployment(ctx, chain2.GetChainReferenceID()),
						).To(BeTrue())
						By("verify that the chain's smart contract id has been deployed", func() {
							ci, err := a.EvmKeeper.GetChainInfo(ctx, chain1.GetChainReferenceID())
							Expect(err).To(BeNil())
							Expect(ci.GetActiveSmartContractID()).To(Equal(smartContract2.GetId()))
						})
					})
				})
			})
		})
	})

	Describe("on snapshot build", func() {

		var snapshot *valsettypes.Snapshot
		When("validator set is valid", func() {
			BeforeEach(func() {
				validators = genValidators(25, 25000)
				for _, val := range validators {
					a.StakingKeeper.SetValidator(ctx, val)
				}
			})

			When("evm chain and smart contract both exist", func() {
				BeforeEach(func() {
					for _, val := range validators {
						private, err := crypto.GenerateKey()
						Expect(err).To(BeNil())
						accAddr := crypto.PubkeyToAddress(private.PublicKey)
						err = a.ValsetKeeper.AddExternalChainInfo(ctx, val.GetOperator(), []*valsettypes.ExternalChainInfo{
							{
								ChainType:        "evm",
								ChainReferenceID: newChain.ChainReferenceID,
								Address:          accAddr.Hex(),
								Pubkey:           []byte("pub key" + accAddr.Hex()),
							},
							{
								ChainType:        "evm",
								ChainReferenceID: "new-chain",
								Address:          accAddr.Hex(),
								Pubkey:           []byte("pub key" + accAddr.Hex()),
							},
						})
						Expect(err).To(BeNil())
					}
					var err error
					snapshot, err = a.ValsetKeeper.TriggerSnapshotBuild(ctx)
					Expect(err).To(BeNil())
				})

				BeforeEach(func() {
					err := a.EvmKeeper.AddSupportForNewChain(
						ctx,
						newChain.GetChainReferenceID(),
						newChain.GetChainID(),
						newChain.GetBlockHeight(),
						newChain.GetBlockHashAtHeight(),
						big.NewInt(55),
					)
					Expect(err).To(BeNil())

					_, err = a.EvmKeeper.SaveNewSmartContract(ctx, smartContract.GetAbiJSON(), smartContract.GetBytecode())
					Expect(err).To(BeNil())

					err = a.EvmKeeper.ActivateChainReferenceID(ctx, newChain.ChainReferenceID, smartContract, "addr", []byte("abc"))
					Expect(err).To(BeNil())

					By("it should have upload smart contract message", func() {
						msgs, err := a.ConsensusKeeper.GetMessagesFromQueue(ctx, "evm/eth-main/evm-turnstone-message", 5)

						Expect(err).To(BeNil())
						Expect(len(msgs)).To(Equal(1))

						con, err := msgs[0].ConsensusMsg(a.AppCodec())
						Expect(err).To(BeNil())

						evmMsg, ok := con.(*types.Message)
						Expect(ok).To(BeTrue())

						_, ok = evmMsg.GetAction().(*types.Message_UploadSmartContract)
						Expect(ok).To(BeTrue())

						a.ConsensusKeeper.DeleteJob(ctx, "evm/eth-main/evm-turnstone-message", msgs[0].GetId())
					})
				})

				It("expects update valset message to exist", func() {
					a.EvmKeeper.OnSnapshotBuilt(ctx, snapshot)
					msgs, err := a.ConsensusKeeper.GetMessagesFromQueue(ctx, "evm/eth-main/evm-turnstone-message", 5)

					Expect(err).To(BeNil())
					Expect(len(msgs)).To(Equal(1))

					con, err := msgs[0].ConsensusMsg(a.AppCodec())
					Expect(err).To(BeNil())

					evmMsg, ok := con.(*types.Message)
					Expect(ok).To(BeTrue())

					_, ok = evmMsg.GetAction().(*types.Message_UpdateValset)
					Expect(ok).To(BeTrue())
				})

				When("adding another chain which is not yet active", func() {
					BeforeEach(func() {
						err := a.EvmKeeper.AddSupportForNewChain(
							ctx,
							"new-chain",
							123,
							uint64(123),
							"0x1234",
							big.NewInt(55),
						)
						Expect(err).To(BeNil())
					})

					It("tries to deploy a smart contract to it", func() {
						a.EvmKeeper.OnSnapshotBuilt(ctx, snapshot)
						msgs, err := a.ConsensusKeeper.GetMessagesFromQueue(ctx, "evm/new-chain/evm-turnstone-message", 5)
						Expect(err).To(BeNil())
						Expect(len(msgs)).To(Equal(1))

						con, err := msgs[0].ConsensusMsg(a.AppCodec())
						Expect(err).To(BeNil())

						evmMsg, ok := con.(*types.Message)
						Expect(ok).To(BeTrue())

						_, ok = evmMsg.GetAction().(*types.Message_UploadSmartContract)
						Expect(ok).To(BeTrue())
					})
				})

				When("there is another upload valset already in", func() {
					BeforeEach(func() {
						err := a.EvmKeeper.AddSupportForNewChain(
							ctx,
							"new-chain",
							123,
							uint64(123),
							"0x1234",
							big.NewInt(55),
						)
						Expect(err).To(BeNil())
						err = a.EvmKeeper.ActivateChainReferenceID(ctx, "new-chain", &types.SmartContract{Id: 123}, "addr", []byte("abc"))
						Expect(err).To(BeNil())
						for _, val := range validators {
							private, err := crypto.GenerateKey()
							Expect(err).To(BeNil())
							accAddr := crypto.PubkeyToAddress(private.PublicKey)
							err = a.ValsetKeeper.AddExternalChainInfo(ctx, val.GetOperator(), []*valsettypes.ExternalChainInfo{
								{
									ChainType:        "evm",
									ChainReferenceID: "new-chain",
									Address:          accAddr.Hex(),
									Pubkey:           []byte("pub key" + accAddr.Hex()),
								},
							})
							Expect(err).To(BeNil())
						}
					})
					BeforeEach(func() {
						msgs, err := a.ConsensusKeeper.GetMessagesFromQueue(ctx, "evm/new-chain/evm-turnstone-message", 5)
						Expect(err).To(BeNil())
						for _, msg := range msgs {
							// we are now clearing the deploy smart contract from the queue as we don't need it
							a.ConsensusKeeper.DeleteJob(ctx, "evm/new-chain/evm-turnstone-message", msg.GetId())
						}
						a.ConsensusKeeper.PutMessageInQueue(ctx, "evm/new-chain/evm-turnstone-message", &types.Message{
							TurnstoneID:      "abc",
							ChainReferenceID: "new-chain",
							Action: &types.Message_UpdateValset{
								UpdateValset: &types.UpdateValset{
									Valset: &types.Valset{
										ValsetID: 777,
									},
								},
							},
						}, nil)
					})
					It("deletes the old smart deployment", func() {
						a.EvmKeeper.OnSnapshotBuilt(ctx, snapshot)
						msgs, err := a.ConsensusKeeper.GetMessagesFromQueue(ctx, "evm/new-chain/evm-turnstone-message", 5)
						Expect(err).To(BeNil())
						Expect(len(msgs)).To(Equal(1))

						con, err := msgs[0].ConsensusMsg(a.AppCodec())
						Expect(err).To(BeNil())

						evmMsg, ok := con.(*types.Message)
						Expect(ok).To(BeTrue())

						vset, ok := evmMsg.GetAction().(*types.Message_UpdateValset)
						Expect(ok).To(BeTrue())
						Expect(vset.UpdateValset.GetValset().GetValsetID()).NotTo(Equal(uint64(777)))
						Expect(len(vset.UpdateValset.GetValset().GetValidators())).NotTo(BeZero())
					})

				})

			})
		})

		When("validator set is too tiny", func() {
			BeforeEach(func() {
				validators = genValidators(25, 25000)[:5]
				for _, val := range validators {
					a.StakingKeeper.SetValidator(ctx, val)
				}
				_, err := a.ValsetKeeper.TriggerSnapshotBuild(ctx)
				Expect(err).To(BeNil())
			})

			Context("evm chain and smart contract both exist", func() {
				BeforeEach(func() {
					err := a.EvmKeeper.AddSupportForNewChain(
						ctx,
						newChain.GetChainReferenceID(),
						newChain.GetChainID(),
						newChain.GetBlockHeight(),
						newChain.GetBlockHashAtHeight(),
						big.NewInt(55),
					)
					Expect(err).To(BeNil())
					_, err = a.EvmKeeper.SaveNewSmartContract(ctx, smartContract.GetAbiJSON(), smartContract.GetBytecode())
					Expect(err).To(BeNil())
				})

				It("doesn't put any message into a queue", func() {
					msgs, err := a.ConsensusKeeper.GetMessagesFromQueue(ctx, "evm/eth-main/evm-turnstone-message", 5)
					Expect(err).To(BeNil())
					Expect(msgs).To(BeZero())
				})
			})

		})
	})
})

var _ = Describe("change min on chain balance", func() {

	var a app.TestApp
	var ctx sdk.Context
	newChain := &types.AddChainProposal{
		ChainReferenceID:  "eth-main",
		Title:             "bla",
		Description:       "bla",
		BlockHeight:       uint64(123),
		BlockHashAtHeight: "0x1234",
	}

	BeforeEach(func() {
		a = app.NewTestApp(GinkgoT(), false)
		ctx = a.NewContext(false, tmproto.Header{
			Height: 5,
		})
	})

	When("chain info exists", func() {
		BeforeEach(func() {
			err := a.EvmKeeper.AddSupportForNewChain(ctx, newChain.GetChainReferenceID(), newChain.GetChainID(), 1, "a", big.NewInt(55))
			Expect(err).To(BeNil())
		})

		BeforeEach(func() {
			ci, err := a.EvmKeeper.GetChainInfo(ctx, newChain.GetChainReferenceID())
			Expect(err).To(BeNil())
			balance, err := ci.GetMinOnChainBalanceBigInt()
			Expect(err).To(BeNil())
			Expect(balance.Text(10)).To(Equal(big.NewInt(55).Text(10)))
		})

		It("changes the on chain balance", func() {
			err := a.EvmKeeper.ChangeMinOnChainBalance(ctx, newChain.GetChainReferenceID(), big.NewInt(888))
			Expect(err).To(BeNil())

			ci, err := a.EvmKeeper.GetChainInfo(ctx, newChain.GetChainReferenceID())
			Expect(err).To(BeNil())
			balance, err := ci.GetMinOnChainBalanceBigInt()
			Expect(err).To(BeNil())
			Expect(balance.Text(10)).To(Equal(big.NewInt(888).Text(10)))
		})
	})

	When("chain info does not exists", func() {
		It("returns an error", func() {
			err := a.EvmKeeper.ChangeMinOnChainBalance(ctx, newChain.GetChainReferenceID(), big.NewInt(888))
			Expect(err).To(MatchError(keeper.ErrChainNotFound))
		})
	})

})
