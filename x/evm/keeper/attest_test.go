package keeper

import (
	"io/ioutil"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethcoretypes "github.com/ethereum/go-ethereum/core/types"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/palomachain/paloma/util/slice"
	consensusmocks "github.com/palomachain/paloma/x/consensus/keeper/consensus/mocks"
	consensustypes "github.com/palomachain/paloma/x/consensus/types"
	valsettypes "github.com/palomachain/paloma/x/valset/types"
	"github.com/stretchr/testify/mock"
	"github.com/vizualni/whoops"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/palomachain/paloma/x/evm/types"
	evmmocks "github.com/palomachain/paloma/x/evm/types/mocks"
)

var (
	contractAbi         = string(whoops.Must(ioutil.ReadFile("testdata/sample-abi.json")))
	contractBytecodeStr = string(whoops.Must(ioutil.ReadFile("testdata/sample-bytecode.out")))

	sampleTx1RawBytes = common.FromHex(string(whoops.Must(ioutil.ReadFile("testdata/sample-tx-raw.hex"))))

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
				consensukeeper.On("PutMessageForSigning", mock.Anything, mock.Anything, mock.Anything).Return(nil)

				err := k.AddSupportForNewChain(
					ctx,
					newChain.GetChainReferenceID(),
					newChain.GetChainID(),
					newChain.GetBlockHeight(),
					newChain.GetBlockHashAtHeight(),
				)
				Expect(err).To(BeNil())

				sc, err := k.SaveNewSmartContract(ctx, contractAbi, common.FromHex(contractBytecodeStr))
				Expect(err).To(BeNil())

				dep, _ := k.getSmartContractDeploying(ctx, sc.GetId(), newChain.GetChainReferenceID())
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
								ChainType:        "EVM",
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
								ChainType:        "EVM",
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
								ChainType:        "EVM",
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
							Proof:      whoops.Must(codectypes.NewAnyWithValue(&types.TxExecutedProof{whoops.Must(sampleTx1.MarshalBinary())})),
						},
						{
							ValAddress: sdk.ValAddress("456"),
							Proof:      whoops.Must(codectypes.NewAnyWithValue(&types.TxExecutedProof{whoops.Must(sampleTx1.MarshalBinary())})),
						},
						{
							ValAddress: sdk.ValAddress("789"),
							Proof:      whoops.Must(codectypes.NewAnyWithValue(&types.TxExecutedProof{whoops.Must(sampleTx1.MarshalBinary())})),
						},
					}
				})
				g.BeforeEach(func() {
					q.On("ChainInfo").Return("", "eth-main")
					q.On("Remove", mock.Anything, uint64(123)).Return(nil)
				})

				successfulProcess := func() {
					g.It("processees it successfully", g.Offset(1), func() {
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
				})

				g.When("message is UpdateValset", func() {
					g.BeforeEach(func() {
						consensusMsg.Action = &types.Message_UpdateValset{
							UpdateValset: &types.UpdateValset{},
						}
					})

					g.BeforeEach(func() {
						q.On("GetAll", mock.Anything).Return(nil, nil)
					})
					successfulProcess()
				})

				g.When("message is UploadSmartContract", func() {
					g.BeforeEach(func() {
						consensusMsg.Action = &types.Message_UploadSmartContract{
							UploadSmartContract: &types.UploadSmartContract{
								Id: 1,
							},
						}
					})

					successfulProcess()

					// g.It("removes the info about smart contract to chain deployment")
				})

				g.JustAfterEach(func() {
					g.By("there is no error when processing evidence")
					Expect(subject()).To(BeNil())
				})

				g.JustAfterEach(func() {
					Expect(k.isTxProcessed(ctx, sampleTx1)).To(BeTrue())
				})

			})

		})

	})

})
