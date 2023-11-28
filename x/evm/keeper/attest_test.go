package keeper

import (
	"errors"
	"math/big"
	"os"
	"sync"

	"github.com/VolumeFi/whoops"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethcoretypes "github.com/ethereum/go-ethereum/core/types"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/palomachain/paloma/util/slice"
	consensusmocks "github.com/palomachain/paloma/x/consensus/keeper/consensus/mocks"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	"github.com/palomachain/paloma/x/evm/types"
	evmmocks "github.com/palomachain/paloma/x/evm/types/mocks"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/mock"
)

var (
	contractAbi         = string(whoops.Must(os.ReadFile("testdata/sample-abi.json")))
	contractBytecodeStr = string(whoops.Must(os.ReadFile("testdata/sample-bytecode.out")))

	sampleTx1RawBytes = common.FromHex(string(whoops.Must(os.ReadFile("testdata/sample-tx-raw.hex"))))

	sampleTx1 = func() *ethcoretypes.Transaction {
		tx := new(ethcoretypes.Transaction)
		whoops.Assert(tx.UnmarshalBinary(sampleTx1RawBytes))
		return tx
	}()
)

var _ = g.Describe("attest router", func() {
	var k Keeper
	var ctx sdk.Context
	var q *consensusmocks.Queuer
	var v *evmmocks.ValsetKeeper
	var consensukeeper *evmmocks.ConsensusKeeper
	var msg *consensustypes.QueuedSignedMessage
	var consensusMsg *types.Message
	var evidence []*consensustypes.Evidence
	var isGoodcase bool
	newChain := &types.AddChainProposal{
		ChainReferenceID:  "eth-main",
		Title:             "bla",
		Description:       "bla",
		BlockHeight:       uint64(123),
		BlockHashAtHeight: "0x1234",
	}

	type valpower struct {
		valAddr       sdk.ValAddress
		power         int64
		externalChain []*valsettypes.ExternalChainInfo
	}
	var valpowers []valpower

	var totalPower int64

	g.BeforeEach(func() {
		t := g.GinkgoT()
		kpr, ms, _ctx := NewEvmKeeper(t)
		ctx = _ctx
		k = *kpr
		v = ms.ValsetKeeper
		consensukeeper = ms.ConsensusKeeper
		q = consensusmocks.NewQueuer(t)
		isGoodcase = true
		ms.GravityKeeper.On("GetLastObservedEventNonce", mock.Anything).Return(uint64(100), nil).Maybe()
	})

	g.BeforeEach(func() {
		consensusMsg = &types.Message{}
	})

	g.JustBeforeEach(func() {
		msg = &consensustypes.QueuedSignedMessage{
			Id:       123,
			Msg:      whoops.Must(codectypes.NewAnyWithValue(consensusMsg)),
			Evidence: evidence,
		}
	})

	var subject func() error
	var subjectOnce sync.Once
	var subjectErr error
	g.BeforeEach(func() {
		subjectOnce = sync.Once{}
		subject = func() error {
			subjectOnce.Do(func() {
				subjectErr = k.attestRouter(ctx, q, msg)
			})
			return subjectErr
		}
	})

	g.When("snapshot returns an error", func() {
		retErr := whoops.String("random error")
		g.BeforeEach(func() {
			v.On("GetCurrentSnapshot", mock.Anything).Return(
				nil,
				retErr,
			)
		})
		g.BeforeEach(func() {
			evidence = []*consensustypes.Evidence{
				{
					Proof: whoops.Must(codectypes.NewAnyWithValue(&types.SmartContractExecutionErrorProof{ErrorMessage: "doesn't matter"})),
				},
				{
					Proof: whoops.Must(codectypes.NewAnyWithValue(&types.SmartContractExecutionErrorProof{ErrorMessage: "just need something to exist"})),
				},
			}
		})

		g.It("returns error back", func() {
			Expect(subject()).To(MatchError(retErr))
		})
	})

	g.When("snapshot returns an actual snapshot", func() {
		g.JustBeforeEach(func() {
			v.On("GetCurrentSnapshot", mock.Anything).Return(
				&valsettypes.Snapshot{
					Validators: slice.Map(valpowers, func(p valpower) valsettypes.Validator {
						return valsettypes.Validator{
							ShareCount:         sdk.NewInt(p.power),
							Address:            p.valAddr,
							ExternalChainInfos: p.externalChain,
						}
					}),
					TotalShares: sdk.NewInt(totalPower),
				},
				nil,
			)
		})

		g.When("there is not enough power to reach a consensus", func() {
			g.BeforeEach(func() {
				totalPower = 20
				valpowers = []valpower{
					{
						valAddr: sdk.ValAddress("123"),
						power:   5,
					},
					{
						valAddr: sdk.ValAddress("456"),
						power:   5,
					},
				}
			})

			g.It("returns nil for error", func() {
				Expect(subject()).To(BeNil())
			})
		})

		g.When("there is enough power to reach a consensus", func() {
			setupChainSupport := func() {
				consensukeeper.On("PutMessageInQueue", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(uint64(10), nil)

				err := k.AddSupportForNewChain(
					ctx,
					newChain.GetChainReferenceID(),
					newChain.GetChainID(),
					newChain.GetBlockHeight(),
					newChain.GetBlockHashAtHeight(),
					big.NewInt(55),
				)
				Expect(err).To(BeNil())

				sc, err := k.SaveNewSmartContract(ctx, contractAbi, common.FromHex(contractBytecodeStr))
				Expect(err).To(BeNil())

				err = k.SetAsCompassContract(ctx, sc)
				Expect(err).To(BeNil())

				dep, _ := k.getSmartContractDeploymentByContractID(ctx, sc.GetId(), newChain.GetChainReferenceID())
				Expect(dep).NotTo(BeNil())
			}

			g.BeforeEach(func() {
				totalPower = 20
				valpowers = []valpower{
					{
						valAddr: sdk.ValAddress("123"),
						power:   5,
						externalChain: []*valsettypes.ExternalChainInfo{
							{
								ChainType:        "evm",
								ChainReferenceID: newChain.GetChainReferenceID(),
								Address:          "addr1",
								Pubkey:           []byte("1"),
							},
						},
					},
					{
						valAddr: sdk.ValAddress("456"),
						power:   5,
						externalChain: []*valsettypes.ExternalChainInfo{
							{
								ChainType:        "evm",
								ChainReferenceID: newChain.GetChainReferenceID(),
								Address:          "addr2",
								Pubkey:           []byte("2"),
							},
						},
					},
					{
						valAddr: sdk.ValAddress("789"),
						power:   5,
						externalChain: []*valsettypes.ExternalChainInfo{
							{
								ChainType:        "evm",
								ChainReferenceID: newChain.GetChainReferenceID(),
								Address:          "addr3",
								Pubkey:           []byte("3"),
							},
						},
					},
				}
			})

			g.Context("with a valid evidence", func() {
				g.BeforeEach(func() {
					evidence = []*consensustypes.Evidence{
						{
							ValAddress: sdk.ValAddress("123"),
							Proof:      whoops.Must(codectypes.NewAnyWithValue(&types.TxExecutedProof{SerializedTX: whoops.Must(sampleTx1.MarshalBinary())})),
						},
						{
							ValAddress: sdk.ValAddress("456"),
							Proof:      whoops.Must(codectypes.NewAnyWithValue(&types.TxExecutedProof{SerializedTX: whoops.Must(sampleTx1.MarshalBinary())})),
						},
						{
							ValAddress: sdk.ValAddress("789"),
							Proof:      whoops.Must(codectypes.NewAnyWithValue(&types.TxExecutedProof{SerializedTX: whoops.Must(sampleTx1.MarshalBinary())})),
						},
					}
				})
				g.BeforeEach(func() {
					q.On("ChainInfo").Return("", "eth-main")
					q.On("Remove", mock.Anything, uint64(123)).Return(nil)
				})

				successfulProcess := func() {
					g.It("processes it successfully", g.Offset(1), func() {
						setupChainSupport()
						Expect(subject()).To(BeNil())
					})
				}

				g.JustBeforeEach(func() {
					Expect(k.isTxProcessed(ctx, sampleTx1)).To(BeFalse())
				})

				g.When("message is SubmitLogicCall", func() {
					g.BeforeEach(func() {
						consensusMsg.Action = &types.Message_SubmitLogicCall{
							SubmitLogicCall: &types.SubmitLogicCall{},
						}
					})
					successfulProcess()

					g.When("message has not been retried", func() {
						g.BeforeEach(func() {
							consensusMsg.Action = &types.Message_SubmitLogicCall{
								SubmitLogicCall: &types.SubmitLogicCall{
									Retries: uint32(0),
								},
							}
							consensukeeper.On("PutMessageInQueue", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(uint64(10), nil)
						})

						g.It("should attempt to retry", func() {
							setupChainSupport()
							Expect(subject()).To(BeNil())
						})
					})

					g.When("message has been retried too many times", func() {
						g.BeforeEach(func() {
							consensusMsg.Action = &types.Message_SubmitLogicCall{
								SubmitLogicCall: &types.SubmitLogicCall{
									Retries: uint32(2),
								},
							}
						})

						g.It("should attempt to retry", func() {
							setupChainSupport()
							Expect(subject()).To(BeNil())
						})
					})
				})

				g.When("message is UpdateValset", func() {
					g.BeforeEach(func() {
						consensusMsg.Action = &types.Message_UpdateValset{
							UpdateValset: &types.UpdateValset{
								Valset: &types.Valset{
									ValsetID: 1,
								},
							},
						}
					})

					g.BeforeEach(func() {
						q.On("GetAll", mock.Anything).Return(nil, nil)
					})

					g.When("successfully sets valset for chain", func() {
						g.BeforeEach(func() {
							v.On("SetSnapshotOnChain", mock.Anything, mock.Anything, mock.Anything).Return(nil)
						})
						successfulProcess()
					})

					g.When("unsuccessfully sets valset for chain", func() {
						// We still process successfully even if we get an error here
						g.BeforeEach(func() {
							v.On("SetSnapshotOnChain", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("example error"))
						})
						successfulProcess()
					})
				})

				g.When("message is UploadSmartContract", func() {
					g.BeforeEach(func() {
						consensusMsg.Action = &types.Message_UploadSmartContract{
							UploadSmartContract: &types.UploadSmartContract{
								Id: 1,
							},
						}
					})

					g.It("updates the info about smart contract to chain deployment", func() {
						setupChainSupport()
						Expect(subject()).To(BeNil())
						v, _ := k.getSmartContractDeploymentByContractID(ctx, uint64(1), newChain.GetChainReferenceID())
						Expect(v.Status).To(Equal(types.SmartContractDeployment_WAITING_FOR_ERC20_OWNERSHIP_TRANSFER))
					})
				})

				g.When("message is TransferERC20Ownership", func() {
					g.BeforeEach(func() {
						consensusMsg.Action = &types.Message_TransferERC20Ownership{
							TransferERC20Ownership: &types.TransferERC20Ownership{
								SmartContractID:   1,
								NewCompassAddress: sdk.ValAddress("12345678901234567890"),
							},
						}
					})

					g.When("contract not actively deploying", func() {
						g.It("returns an error", func() {
							isGoodcase = false
							setupChainSupport()
							Expect(subject()).To(MatchError("trying to activate a smart contract that is not currently deploying"))
						})
					})

					g.When("transfer is done", func() {
						g.It("must remove the deployment from deployment store", func() {
							setupChainSupport()
							k.SetSmartContractDeploymentStatusByContractID(ctx, uint64(1), newChain.ChainReferenceID, types.SmartContractDeployment_WAITING_FOR_ERC20_OWNERSHIP_TRANSFER)
							Expect(subject()).To(BeNil())
							v, _ := k.getSmartContractDeploymentByContractID(ctx, uint64(1), newChain.GetChainReferenceID())
							Expect(v).To(BeNil())
						})
					})
				})

				g.JustAfterEach(func() {
					if isGoodcase {
						g.By("there is no error when processing evidence")
						Expect(subject()).To(BeNil())
					} else {
						g.By("there is an error when processing evidence")
						Expect(subject()).To(Not(BeNil()))
					}
				})

				g.JustAfterEach(func() {
					Expect(k.isTxProcessed(ctx, sampleTx1)).To(Equal(isGoodcase))
				})
			})
		})
	})
})
