package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/palomachain/paloma/x/gravity/types"
)

// nolint: exhaustruct
func TestQueryValsetConfirm(t *testing.T) {
	var (
		addrStr                       = "gravity1ees2tqhhhm9ahlhceh2zdguww9lqn2ckcxpllh"
		nonce                         = uint64(1)
		myValidatorCosmosAddr, err1   = sdk.AccAddressFromBech32(addrStr)
		myValidatorEthereumAddr, err2 = types.NewEthAddress("0x3232323232323232323232323232323232323232")
	)
	require.NoError(t, err1)
	require.NoError(t, err2)
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	sdkCtx := input.Context
	ctx := sdk.WrapSDKContext(input.Context)
	k := input.GravityKeeper
	input.GravityKeeper.SetValsetConfirm(sdkCtx, types.MsgValsetConfirm{
		Nonce:        nonce,
		Orchestrator: myValidatorCosmosAddr.String(),
		EthAddress:   myValidatorEthereumAddr.GetAddress().Hex(),
		Signature:    "abcdef123456789",
	})

	specs := map[string]struct {
		src     types.QueryValsetConfirmRequest
		expErr  bool
		expResp types.QueryValsetConfirmResponse
	}{
		/*  Nonce        uint64 `protobuf:"varint,1,opt,name=nonce,proto3" json:"nonce,omitempty"`
		    Orchestrator string `protobuf:"bytes,2,opt,name=orchestrator,proto3" json:"orchestrator,omitempty"`
		    EthAddress   string `protobuf:"bytes,3,opt,name=eth_address,json=ethAddress,proto3" json:"eth_address,omitempty"`
		    Signature    string `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty"`
		}*/

		"all good": {
			src: types.QueryValsetConfirmRequest{Nonce: 1, Address: myValidatorCosmosAddr.String()},

			// expResp:  []byte(`{"type":"gravity/MsgValsetConfirm", "value":{"eth_address":"0x3232323232323232323232323232323232323232", "nonce": "1", "orchestrator": "cosmos1ees2tqhhhm9ahlhceh2zdguww9lqn2ckukn86l",  "signature": "alksdjhflkasjdfoiasjdfiasjdfoiasdj"}}`),
			expResp: types.QueryValsetConfirmResponse{
				Confirm: types.NewMsgValsetConfirm(1, *myValidatorEthereumAddr, myValidatorCosmosAddr, "abcdef123456789")},
			expErr: false,
		},
		"unknown nonce": {
			src:     types.QueryValsetConfirmRequest{Nonce: 999999, Address: myValidatorCosmosAddr.String()},
			expResp: types.QueryValsetConfirmResponse{Confirm: nil},
		},
		"invalid address": {
			src:    types.QueryValsetConfirmRequest{Nonce: 1, Address: "not a valid addr"},
			expErr: true,
		},
	}

	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			got, err := k.ValsetConfirm(ctx, &spec.src)
			if spec.expErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if spec.expResp == (types.QueryValsetConfirmResponse{}) {
				assert.True(t, got == nil || got.Confirm == nil)
				return
			}
			assert.Equal(t, &spec.expResp, got)
		})
	}
}

// nolint: exhaustruct
func TestAllValsetConfirmsBynonce(t *testing.T) {
	addrs := []string{
		"gravity1u508cfnsk2nhakv80vdtq3nf558ngyvlfxm2hd",
		"gravity1krtcsrxhadj54px0vy6j33pjuzcd3jj8jtz98y",
		"gravity1u94xef3cp9thkcpxecuvhtpwnmg8mhljeh96n9",
	}
	var (
		nonce                        = uint64(1)
		myValidatorCosmosAddr1, e1   = sdk.AccAddressFromBech32(addrs[0])
		myValidatorCosmosAddr2, e2   = sdk.AccAddressFromBech32(addrs[1])
		myValidatorCosmosAddr3, e3   = sdk.AccAddressFromBech32(addrs[2])
		myValidatorEthereumAddr1, e4 = types.NewEthAddress("0x0101010101010101010101010101010101010101")
		myValidatorEthereumAddr2, e5 = types.NewEthAddress("0x0202020202020202020202020202020202020202")
		myValidatorEthereumAddr3, e6 = types.NewEthAddress("0x0303030303030303030303030303030303030303")
	)
	require.NoError(t, e1)
	require.NoError(t, e2)
	require.NoError(t, e3)
	require.NoError(t, e4)
	require.NoError(t, e5)
	require.NoError(t, e6)

	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	sdkCtx := input.Context
	ctx := sdk.WrapSDKContext(input.Context)
	k := input.GravityKeeper

	// seed confirmations
	for i := 0; i < 3; i++ {
		addr, err := sdk.AccAddressFromBech32(addrs[i])
		require.NoError(t, err)
		msg := types.MsgValsetConfirm{}
		msg.EthAddress = gethcommon.BytesToAddress(bytes.Repeat([]byte{byte(i + 1)}, 20)).String()
		msg.Nonce = uint64(1)
		msg.Orchestrator = addr.String()
		msg.Signature = fmt.Sprintf("d34db33f%d", i)
		input.GravityKeeper.SetValsetConfirm(sdkCtx, msg)
	}

	specs := map[string]struct {
		src     types.QueryValsetConfirmsByNonceRequest
		expErr  bool
		expResp types.QueryValsetConfirmsByNonceResponse
	}{
		"all good": {
			src: types.QueryValsetConfirmsByNonceRequest{Nonce: 1},
			expResp: types.QueryValsetConfirmsByNonceResponse{Confirms: []types.MsgValsetConfirm{
				*types.NewMsgValsetConfirm(nonce, *myValidatorEthereumAddr2, myValidatorCosmosAddr2, "d34db33f1"),
				*types.NewMsgValsetConfirm(nonce, *myValidatorEthereumAddr3, myValidatorCosmosAddr3, "d34db33f2"),
				*types.NewMsgValsetConfirm(nonce, *myValidatorEthereumAddr1, myValidatorCosmosAddr1, "d34db33f0"),
			}},
		},
		"unknown nonce": {
			src:     types.QueryValsetConfirmsByNonceRequest{Nonce: 999999},
			expResp: types.QueryValsetConfirmsByNonceResponse{},
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			got, err := k.ValsetConfirmsByNonce(ctx, &types.QueryValsetConfirmsByNonceRequest{Nonce: spec.src.Nonce})
			if spec.expErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, &(spec.expResp), got)
		})
	}
}

// TODO: Check failure modes
// nolint: exhaustruct
func TestLastValsetRequests(t *testing.T) {
	val1 := types.Valset{
		Nonce:        6,
		Height:       1235167,
		RewardAmount: sdk.ZeroInt(),
		RewardToken:  "0x0000000000000000000000000000000000000000",
		Members: []types.BridgeValidator{
			{
				Power:           858993459,
				EthereumAddress: "0x0000000000000000000000000000000000000000",
			},
			{
				Power:           858993459,
				EthereumAddress: "0x0101010101010101010101010101010101010101",
			},
			{
				Power:           858993459,
				EthereumAddress: "0x0202020202020202020202020202020202020202",
			},
			{
				Power:           858993459,
				EthereumAddress: "0x0303030303030303030303030303030303030303",
			},
			{
				Power:           858993459,
				EthereumAddress: "0x0404040404040404040404040404040404040404",
			},
		},
	}

	val2 := types.Valset{
		Nonce:        5,
		Height:       1235067,
		RewardAmount: sdk.ZeroInt(),
		RewardToken:  "0x0000000000000000000000000000000000000000",
		Members: []types.BridgeValidator{
			{
				Power:           858993459,
				EthereumAddress: "0x0000000000000000000000000000000000000000",
			},
			{
				Power:           858993459,
				EthereumAddress: "0x0101010101010101010101010101010101010101",
			},
			{
				Power:           858993459,
				EthereumAddress: "0x0202020202020202020202020202020202020202",
			},
			{
				Power:           858993459,
				EthereumAddress: "0x0303030303030303030303030303030303030303",
			},
			{
				Power:           858993459,
				EthereumAddress: "0x0404040404040404040404040404040404040404",
			},
		},
	}

	val3 := types.Valset{
		Nonce:        4,
		Height:       1234967,
		RewardAmount: sdk.ZeroInt(),
		RewardToken:  "0x0000000000000000000000000000000000000000",
		Members: []types.BridgeValidator{
			{
				Power:           1073741824,
				EthereumAddress: "0x0000000000000000000000000000000000000000",
			},
			{
				Power:           1073741824,
				EthereumAddress: "0x0101010101010101010101010101010101010101",
			},
			{
				Power:           1073741824,
				EthereumAddress: "0x0202020202020202020202020202020202020202",
			},
			{
				Power:           1073741824,
				EthereumAddress: "0x0303030303030303030303030303030303030303",
			},
		},
	}

	val4 := types.Valset{
		Nonce:        3,
		Height:       1234867,
		RewardAmount: sdk.ZeroInt(),
		RewardToken:  "0x0000000000000000000000000000000000000000",
		Members: []types.BridgeValidator{
			{
				Power:           1431655765,
				EthereumAddress: "0x0000000000000000000000000000000000000000",
			},
			{
				Power:           1431655765,
				EthereumAddress: "0x0101010101010101010101010101010101010101",
			},
			{
				Power:           1431655765,
				EthereumAddress: "0x0202020202020202020202020202020202020202",
			},
		},
	}

	val5 := types.Valset{
		Nonce:        2,
		Height:       1234767,
		RewardAmount: sdk.ZeroInt(),
		RewardToken:  "0x0000000000000000000000000000000000000000",
		Members: []types.BridgeValidator{
			{
				Power:           2147483648,
				EthereumAddress: "0x0000000000000000000000000000000000000000",
			},
			{
				Power:           2147483648,
				EthereumAddress: "0x0101010101010101010101010101010101010101",
			},
		},
	}

	valArray := &types.Valsets{val1, val2, val3, val4, val5}

	specs := map[string]struct {
		expResp types.QueryLastValsetRequestsResponse
	}{ // Expect only maxValsetRequestsReturns back
		"limit at 5": {
			expResp: types.QueryLastValsetRequestsResponse{Valsets: *valArray},
		},
	}
	// any lower than this and a validator won't be created
	const minStake = 1000000
	input, _ := SetupTestChain(t, []uint64{minStake, minStake, minStake, minStake, minStake}, true)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := sdk.WrapSDKContext(input.Context)

	// one more valset request

	// increase block height by 100 blocks
	input.Context = input.Context.WithBlockHeight(input.Context.BlockHeight() + 100)

	// Run the staking endblocker to ensure valset is correct in state
	staking.EndBlocker(input.Context, &input.StakingKeeper)

	input.GravityKeeper.SetValsetRequest(input.Context)

	k := input.GravityKeeper
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			got, err := k.LastValsetRequests(ctx, &types.QueryLastValsetRequestsRequest{})
			require.NoError(t, err)
			assert.Equal(t, &spec.expResp, got)
		})
	}
}

// nolint: exhaustruct
// TODO: check that it doesn't accidently return a valset that HAS been signed
// Right now it is basically just testing that any valset comes back
func TestPendingValsetRequests(t *testing.T) {
	specs := map[string]struct {
		expResp types.QueryLastPendingValsetRequestByAddrResponse
	}{
		"find valset": {
			expResp: types.QueryLastPendingValsetRequestByAddrResponse{Valsets: []types.Valset{
				{
					Nonce:        6,
					Height:       1235167,
					RewardAmount: sdk.ZeroInt(),
					RewardToken:  "0x0000000000000000000000000000000000000000",
					Members: []types.BridgeValidator{
						{
							Power:           858993459,
							EthereumAddress: "0x0000000000000000000000000000000000000000",
						},
						{
							Power:           858993459,
							EthereumAddress: "0x0101010101010101010101010101010101010101",
						},
						{
							Power:           858993459,
							EthereumAddress: "0x0202020202020202020202020202020202020202",
						},
						{
							Power:           858993459,
							EthereumAddress: "0x0303030303030303030303030303030303030303",
						},
						{
							Power:           858993459,
							EthereumAddress: "0x0404040404040404040404040404040404040404",
						},
					},
				},
				{
					Nonce:        5,
					Height:       1235067,
					RewardAmount: sdk.ZeroInt(),
					RewardToken:  "0x0000000000000000000000000000000000000000",
					Members: []types.BridgeValidator{
						{
							Power:           858993459,
							EthereumAddress: "0x0000000000000000000000000000000000000000",
						},
						{
							Power:           858993459,
							EthereumAddress: "0x0101010101010101010101010101010101010101",
						},
						{
							Power:           858993459,
							EthereumAddress: "0x0202020202020202020202020202020202020202",
						},
						{
							Power:           858993459,
							EthereumAddress: "0x0303030303030303030303030303030303030303",
						},
						{
							Power:           858993459,
							EthereumAddress: "0x0404040404040404040404040404040404040404",
						},
					},
				},
				{
					Nonce:        4,
					Height:       1234967,
					RewardAmount: sdk.ZeroInt(),
					RewardToken:  "0x0000000000000000000000000000000000000000",
					Members: []types.BridgeValidator{
						{
							Power:           1073741824,
							EthereumAddress: "0x0000000000000000000000000000000000000000",
						},
						{
							Power:           1073741824,
							EthereumAddress: "0x0101010101010101010101010101010101010101",
						},
						{
							Power:           1073741824,
							EthereumAddress: "0x0202020202020202020202020202020202020202",
						},
						{
							Power:           1073741824,
							EthereumAddress: "0x0303030303030303030303030303030303030303",
						},
					},
				},
				{
					Nonce:        3,
					Height:       1234867,
					RewardAmount: sdk.ZeroInt(),
					RewardToken:  "0x0000000000000000000000000000000000000000",
					Members: []types.BridgeValidator{
						{
							Power:           1431655765,
							EthereumAddress: "0x0000000000000000000000000000000000000000",
						},
						{
							Power:           1431655765,
							EthereumAddress: "0x0101010101010101010101010101010101010101",
						},
						{
							Power:           1431655765,
							EthereumAddress: "0x0202020202020202020202020202020202020202",
						},
					},
				},
				{
					Nonce:        2,
					Height:       1234767,
					RewardAmount: sdk.ZeroInt(),
					RewardToken:  "0x0000000000000000000000000000000000000000",
					Members: []types.BridgeValidator{
						{
							Power:           2147483648,
							EthereumAddress: "0x0000000000000000000000000000000000000000",
						},
						{
							Power:           2147483648,
							EthereumAddress: "0x0101010101010101010101010101010101010101",
						},
					},
				},
				{
					Nonce: 1,
					Members: []types.BridgeValidator{
						{
							Power:           4294967296,
							EthereumAddress: "0x0000000000000000000000000000000000000000",
						},
					},
					Height:       1234667,
					RewardAmount: sdk.NewInt(0),
					RewardToken:  "0x0000000000000000000000000000000000000000",
				},
			},
			},
		},
	}
	// any lower than this and a validator won't be created
	const minStake = 1000000
	input, _ := SetupTestChain(t, []uint64{minStake, minStake, minStake, minStake, minStake}, true)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := sdk.WrapSDKContext(input.Context)

	// one more valset request

	// increase block height by 100 blocks
	input.Context = input.Context.WithBlockHeight(input.Context.BlockHeight() + 100)

	// Run the staking endblocker to ensure valset is correct in state
	staking.EndBlocker(input.Context, &input.StakingKeeper)

	input.GravityKeeper.SetValsetRequest(input.Context)

	var valAddr sdk.AccAddress = bytes.Repeat([]byte{byte(1)}, 20)
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			req := new(types.QueryLastPendingValsetRequestByAddrRequest)
			req.Address = valAddr.String()
			got, err := input.GravityKeeper.LastPendingValsetRequestByAddr(ctx, req)
			require.NoError(t, err)
			assert.Equal(t, &spec.expResp, got, got)
		})
	}
}

// nolint: exhaustruct
// TODO: check that it actually returns a batch that has NOT been signed, not just any batch
func TestLastPendingBatchRequest(t *testing.T) {

	specs := map[string]struct {
		expResp types.QueryLastPendingBatchRequestByAddrResponse
	}{
		"find batch": {
			expResp: types.QueryLastPendingBatchRequestByAddrResponse{Batch: []types.OutgoingTxBatch{
				{
					BatchNonce:   1,
					BatchTimeout: 0,
					Transactions: []types.OutgoingTransferTx{
						{
							Id:          2,
							Sender:      "gravity1qyqszqgpqyqszqgpqyqszqgpqyqszqgpkrnxg5",
							DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
							Erc20Token: types.ERC20Token{
								Amount:   sdk.NewInt(101),
								Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
							},
							Erc20Fee: types.ERC20Token{
								Amount:   sdk.NewInt(3),
								Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
							},
						},
						{
							Id:          3,
							Sender:      "gravity1qyqszqgpqyqszqgpqyqszqgpqyqszqgpkrnxg5",
							DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
							Erc20Token: types.ERC20Token{
								Amount:   sdk.NewInt(102),
								Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
							},
							Erc20Fee: types.ERC20Token{
								Amount:   sdk.NewInt(2),
								Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
							},
						},
					},
					TokenContract:      "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
					CosmosBlockCreated: 1235067,
				},
			},
			},
		},
	}
	// any lower than this and a validator won't be created
	const minStake = 1000000
	input, _ := SetupTestChain(t, []uint64{minStake, minStake, minStake, minStake, minStake}, true)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := sdk.WrapSDKContext(input.Context)
	var valAddr sdk.AccAddress = bytes.Repeat([]byte{byte(1)}, 20)
	createTestBatch(t, input, 2)
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			req := new(types.QueryLastPendingBatchRequestByAddrRequest)
			req.Address = valAddr.String()
			got, err := input.GravityKeeper.LastPendingBatchRequestByAddr(ctx, req)
			require.NoError(t, err)
			assert.Equal(t, &spec.expResp, got, got)
		})
	}
}

// nolint: exhaustruct
func createTestBatch(t *testing.T, input TestInput, maxTxElements uint) {
	var (
		mySender            = bytes.Repeat([]byte{1}, 20)
		myReceiver          = "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934"
		myTokenContractAddr = "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B"
		now                 = time.Now().UTC()
	)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	tokenContract, err := types.NewEthAddress(myTokenContractAddr)
	require.NoError(t, err)
	// mint some voucher first
	token, err := types.NewInternalERC20Token(sdk.NewInt(99999), myTokenContractAddr)
	require.NoError(t, err)
	allVouchers := sdk.Coins{token.GravityCoin()}
	err = input.BankKeeper.MintCoins(input.Context, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(input.Context, mySender)
	err = input.BankKeeper.SendCoinsFromModuleToAccount(input.Context, types.ModuleName, mySender, allVouchers)
	require.NoError(t, err)

	// add some TX to the pool
	for i, v := range []uint64{2, 3, 2, 1} {
		amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr)
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr)
		require.NoError(t, err)
		fee := feeToken.GravityCoin()
		_, err = input.GravityKeeper.AddToOutgoingPool(input.Context, mySender, *receiver, amount, fee)
		require.NoError(t, err)
		// Should create:
		// 1: amount 100, fee 2
		// 2: amount 101, fee 3
		// 3: amount 102, fee 2
		// 4: amount 103, fee 1
	}
	// when
	input.Context = input.Context.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	_, err = input.GravityKeeper.BuildOutgoingTXBatch(input.Context, *tokenContract, maxTxElements)
	require.NoError(t, err)
	// Should have 2 and 3 from above
	// 1 and 4 should be unbatched
}

// nolint: exhaustruct
func TestQueryAllBatchConfirms(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	sdkCtx := input.Context
	ctx := sdk.WrapSDKContext(input.Context)
	k := input.GravityKeeper

	var (
		tokenContract      = "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B"
		validatorAddr, err = sdk.AccAddressFromBech32("gravity1mgamdcs9dah0vn0gqupl05up7pedg2mvc3tzjl")
	)
	require.NoError(t, err)

	input.GravityKeeper.SetBatchConfirm(sdkCtx, &types.MsgConfirmBatch{
		Nonce:         1,
		TokenContract: tokenContract,
		EthSigner:     "0xf35e2cc8e6523d683ed44870f5b7cc785051a77d",
		Orchestrator:  validatorAddr.String(),
		Signature:     "d34db33f",
	})

	batchConfirms, err := k.BatchConfirms(ctx, &types.QueryBatchConfirmsRequest{Nonce: 1, ContractAddress: tokenContract})
	require.NoError(t, err)

	expectedRes := types.QueryBatchConfirmsResponse{
		Confirms: []types.MsgConfirmBatch{
			{
				Nonce:         1,
				TokenContract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
				EthSigner:     "0xf35e2cc8e6523d683ed44870f5b7cc785051a77d",
				Orchestrator:  "gravity1mgamdcs9dah0vn0gqupl05up7pedg2mvc3tzjl",
				Signature:     "d34db33f",
			},
		},
	}

	assert.Equal(t, &expectedRes, batchConfirms, "json is equal")
}

// nolint: exhaustruct
func TestQueryLogicCalls(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	sdkCtx := input.Context
	ctx := sdk.WrapSDKContext(input.Context)
	k := input.GravityKeeper
	var (
		logicContract            = "0x510ab76899430424d209a6c9a5b9951fb8a6f47d"
		payload                  = []byte("fake bytes")
		tokenContract            = "0x7580bfe88dd3d07947908fae12d95872a260f2d8"
		invalidationId           = []byte("GravityTesting")
		invalidationNonce uint64 = 1
	)

	// seed with valset requests and eth addresses to make validators
	// that we will later use to lookup calls to be signed
	for i := 0; i < 6; i++ {
		var validators []sdk.ValAddress
		fmt.Printf("%v", validators)
		for j := 0; j <= i; j++ {
			// add an validator each block
			// TODO: replace with real SDK addresses
			valAddr := bytes.Repeat([]byte{byte(j)}, 20)
			ethAddr, err := types.NewEthAddress(gethcommon.BytesToAddress(bytes.Repeat([]byte{byte(j + 1)}, 20)).String())
			require.NoError(t, err)
			input.GravityKeeper.SetEthAddressForValidator(sdkCtx, valAddr, *ethAddr)
			validators = append(validators, valAddr)
		}
	}

	token := []types.ERC20Token{{
		Contract: tokenContract,
		Amount:   sdk.NewIntFromUint64(5000),
	}}

	call := types.OutgoingLogicCall{
		Transfers:            token,
		Fees:                 token,
		LogicContractAddress: logicContract,
		Payload:              payload,
		Timeout:              10000,
		InvalidationId:       invalidationId,
		InvalidationNonce:    uint64(invalidationNonce),
	}
	k.SetOutgoingLogicCall(sdkCtx, call)

	res := k.GetOutgoingLogicCall(sdkCtx, invalidationId, invalidationNonce)

	require.Equal(t, call, *res)

	_, err := k.OutgoingLogicCalls(ctx, &types.QueryOutgoingLogicCallsRequest{})
	require.NoError(t, err)

	var valAddr sdk.AccAddress = bytes.Repeat([]byte{byte(1)}, 20)
	_, err = k.LastPendingLogicCallByAddr(ctx, &types.QueryLastPendingLogicCallByAddrRequest{Address: valAddr.String()})
	require.NoError(t, err)

	require.NoError(t, err)
}

// nolint: exhaustruct
func TestQueryLogicCallConfirms(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	sdkCtx := input.Context
	k := input.GravityKeeper
	var (
		logicContract            = "0x510ab76899430424d209a6c9a5b9951fb8a6f47d"
		payload                  = []byte("fake bytes")
		tokenContract            = "0x7580bfe88dd3d07947908fae12d95872a260f2d8"
		invalidationId           = []byte("GravityTesting")
		invalidationNonce uint64 = 1
	)

	// seed with valset requests and eth addresses to make validators
	// that we will later use to lookup calls to be signed
	for i := 0; i < 6; i++ {
		var validators []sdk.ValAddress
		fmt.Printf("%v", validators)
		for j := 0; j <= i; j++ {
			// add an validator each block
			// TODO: replace with real SDK addresses
			valAddr := bytes.Repeat([]byte{byte(j)}, 20)
			ethAddr, err := types.NewEthAddress(gethcommon.BytesToAddress(bytes.Repeat([]byte{byte(j + 1)}, 20)).String())
			require.NoError(t, err)
			input.GravityKeeper.SetEthAddressForValidator(sdkCtx, valAddr, *ethAddr)
			validators = append(validators, valAddr)
		}
	}

	token := []types.ERC20Token{{
		Contract: tokenContract,
		Amount:   sdk.NewIntFromUint64(5000),
	}}

	call := types.OutgoingLogicCall{
		Transfers:            token,
		Fees:                 token,
		LogicContractAddress: logicContract,
		Payload:              payload,
		Timeout:              10000,
		InvalidationId:       invalidationId,
		InvalidationNonce:    uint64(invalidationNonce),
	}
	k.SetOutgoingLogicCall(sdkCtx, call)

	var valAddr sdk.AccAddress = bytes.Repeat([]byte{byte(1)}, 20)

	ethSigner := gethcommon.BytesToAddress(bytes.Repeat([]byte{0x1}, 20)).String()
	confirm := types.MsgConfirmLogicCall{
		InvalidationId:    hex.EncodeToString(invalidationId),
		InvalidationNonce: 1,
		EthSigner:         ethSigner,
		Orchestrator:      valAddr.String(),
		Signature:         "d34db33f",
	}

	k.SetLogicCallConfirm(sdkCtx, &confirm)

	res := k.GetLogicConfirmsByInvalidationIdAndNonce(sdkCtx, invalidationId, 1)
	assert.Equal(t, len(res), 1)
}

// nolint: exhaustruct
// TODO: test that it gets the correct batch, not just any batch.
// Check with multiple nonces and tokenContracts
func TestQueryBatch(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := sdk.WrapSDKContext(input.Context)
	k := input.GravityKeeper

	var (
		tokenContract = "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B"
	)

	createTestBatch(t, input, 2)

	batch, err := k.BatchRequestByNonce(ctx, &types.QueryBatchRequestByNonceRequest{Nonce: 1, ContractAddress: tokenContract})
	require.NoError(t, err)

	expectedRes := types.QueryBatchRequestByNonceResponse{
		Batch: types.OutgoingTxBatch{
			BatchTimeout: 0,
			Transactions: []types.OutgoingTransferTx{
				{
					Erc20Fee: types.ERC20Token{
						Amount:   sdk.NewInt(3),
						Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
					},
					DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
					Erc20Token: types.ERC20Token{
						Amount:   sdk.NewInt(101),
						Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
					},
					Sender: "gravity1qyqszqgpqyqszqgpqyqszqgpqyqszqgpkrnxg5",
					Id:     2,
				},
				{
					Erc20Fee: types.ERC20Token{
						Amount:   sdk.NewInt(2),
						Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
					},
					DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
					Erc20Token: types.ERC20Token{
						Amount:   sdk.NewInt(102),
						Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
					},
					Sender: "gravity1qyqszqgpqyqszqgpqyqszqgpqyqszqgpkrnxg5",
					Id:     3,
				},
			},
			BatchNonce:         1,
			CosmosBlockCreated: 1234567,
			TokenContract:      "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
		},
	}

	// TODO: this test is failing on the empty representation of valset members
	assert.Equal(t, &expectedRes, batch, batch)
}

// nolint: exhaustruct
func TestLastBatchesRequest(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	ctx := sdk.WrapSDKContext(input.Context)
	k := input.GravityKeeper

	createTestBatch(t, input, 2)
	createTestBatch(t, input, 3)

	lastBatches, err := k.OutgoingTxBatches(ctx, &types.QueryOutgoingTxBatchesRequest{})
	require.NoError(t, err)

	expectedRes := types.QueryOutgoingTxBatchesResponse{
		Batches: []types.OutgoingTxBatch{
			{
				BatchTimeout: 0,
				Transactions: []types.OutgoingTransferTx{
					{
						Erc20Fee: types.ERC20Token{
							Amount:   sdk.NewInt(3),
							Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
						},
						DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
						Erc20Token: types.ERC20Token{
							Amount:   sdk.NewInt(101),
							Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
						},
						Sender: "gravity1qyqszqgpqyqszqgpqyqszqgpqyqszqgpkrnxg5",
						Id:     6,
					},
					{
						Erc20Fee: types.ERC20Token{
							Amount:   sdk.NewInt(2),
							Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
						},
						DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
						Erc20Token: types.ERC20Token{
							Amount:   sdk.NewInt(102),
							Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
						},
						Sender: "gravity1qyqszqgpqyqszqgpqyqszqgpqyqszqgpkrnxg5",
						Id:     7,
					},
					{
						Erc20Fee: types.ERC20Token{
							Amount:   sdk.NewInt(2),
							Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
						},
						DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
						Erc20Token: types.ERC20Token{
							Amount:   sdk.NewInt(100),
							Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
						},
						Sender: "gravity1qyqszqgpqyqszqgpqyqszqgpqyqszqgpkrnxg5",
						Id:     5,
					},
				},
				BatchNonce:         2,
				CosmosBlockCreated: 1234567,
				TokenContract:      "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
			},
			{
				BatchTimeout: 0,
				Transactions: []types.OutgoingTransferTx{
					{
						Erc20Fee: types.ERC20Token{
							Amount:   sdk.NewInt(3),
							Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
						},
						DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
						Erc20Token: types.ERC20Token{
							Amount:   sdk.NewInt(101),
							Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
						},
						Sender: "gravity1qyqszqgpqyqszqgpqyqszqgpqyqszqgpkrnxg5",
						Id:     2,
					},
					{
						Erc20Fee: types.ERC20Token{
							Amount:   sdk.NewInt(2),
							Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
						},
						DestAddress: "0x320915BD0F1bad11cBf06e85D5199DBcAC4E9934",
						Erc20Token: types.ERC20Token{
							Amount:   sdk.NewInt(102),
							Contract: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
						},
						Sender: "gravity1qyqszqgpqyqszqgpqyqszqgpqyqszqgpkrnxg5",
						Id:     3,
					},
				},
				BatchNonce:         1,
				CosmosBlockCreated: 1234567,
				TokenContract:      "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
			},
		},
	}

	assert.Equal(t, &expectedRes, lastBatches, "json is equal")
}

// nolint: exhaustruct
// tests setting and querying eth address and orchestrator addresses
func TestQueryCurrentValset(t *testing.T) {
	var (
		expectedValset = types.Valset{
			Nonce:        1,
			Height:       1234567,
			RewardAmount: sdk.ZeroInt(),
			RewardToken:  "0x0000000000000000000000000000000000000000",
			Members: []types.BridgeValidator{
				{
					Power:           858993459,
					EthereumAddress: "0x0101010101010101010101010101010101010101",
				},
				{
					Power:           858993459,
					EthereumAddress: "0x0202020202020202020202020202020202020202",
				},
				{
					Power:           858993459,
					EthereumAddress: "0x0303030303030303030303030303030303030303",
				},
				{
					Power:           858993459,
					EthereumAddress: "0x0404040404040404040404040404040404040404",
				},
				{
					Power:           858993459,
					EthereumAddress: "0x0505050505050505050505050505050505050505",
				},
			},
		}
	)
	input, _ := SetupFiveValChain(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	sdkCtx := input.Context

	currentValset, err := input.GravityKeeper.GetCurrentValset(sdkCtx)
	require.NoError(t, err)

	assert.Equal(t, expectedValset, currentValset)
}

// nolint: exhaustruct
func TestQueryERC20ToDenom(t *testing.T) {
	var (
		erc20, err = types.NewEthAddress("0xb462864E395d88d6bc7C5dd5F3F5eb4cc2599255")
		denom      = "uatom"
	)
	require.NoError(t, err)
	response := types.QueryERC20ToDenomResponse{
		Denom:            denom,
		CosmosOriginated: true,
	}
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	sdkCtx := input.Context
	ctx := sdk.WrapSDKContext(input.Context)
	k := input.GravityKeeper
	input.GravityKeeper.setCosmosOriginatedDenomToERC20(sdkCtx, denom, *erc20)

	queriedDenom, err := k.ERC20ToDenom(ctx, &types.QueryERC20ToDenomRequest{Erc20: erc20.GetAddress().Hex()})
	require.NoError(t, err)

	assert.Equal(t, &response, queriedDenom)
}

// nolint: exhaustruct
func TestQueryDenomToERC20(t *testing.T) {
	var (
		erc20, err = types.NewEthAddress("0xb462864E395d88d6bc7C5dd5F3F5eb4cc2599255")
		denom      = "uatom"
	)
	require.NoError(t, err)
	response := types.QueryDenomToERC20Response{
		Erc20:            erc20.GetAddress().Hex(),
		CosmosOriginated: true,
	}
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	sdkCtx := input.Context
	ctx := sdk.WrapSDKContext(input.Context)
	k := input.GravityKeeper
	input.GravityKeeper.setCosmosOriginatedDenomToERC20(sdkCtx, denom, *erc20)

	queriedERC20, err := k.DenomToERC20(ctx, &types.QueryDenomToERC20Request{Denom: denom})
	require.NoError(t, err)

	assert.Equal(t, &response, queriedERC20)
}

// nolint: exhaustruct
func TestQueryPendingSendToEth(t *testing.T) {
	input := CreateTestEnv(t)
	defer func() { input.Context.Logger().Info("Asserting invariants at test end"); input.AssertInvariants() }()

	sdkCtx := input.Context
	ctx := sdk.WrapSDKContext(input.Context)
	k := input.GravityKeeper
	var (
		now                 = time.Now().UTC()
		mySender, err1      = sdk.AccAddressFromBech32("gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm")
		myReceiver          = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
		myTokenContractAddr = "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5" // Pickle
		token, err2         = types.NewInternalERC20Token(sdk.NewInt(99999), myTokenContractAddr)
		allVouchers         = sdk.NewCoins(token.GravityCoin())
	)
	require.NoError(t, err1)
	require.NoError(t, err2)
	receiver, err := types.NewEthAddress(myReceiver)
	require.NoError(t, err)
	tokenContract, err := types.NewEthAddress(myTokenContractAddr)
	require.NoError(t, err)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(sdkCtx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(sdkCtx, mySender)
	require.NoError(t, input.BankKeeper.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, mySender, allVouchers))

	// CREATE FIRST BATCH
	// ==================

	// add some TX to the pool
	for i, v := range []uint64{2, 3, 2, 1} {
		amountToken, err := types.NewInternalERC20Token(sdk.NewInt(int64(i+100)), myTokenContractAddr)
		require.NoError(t, err)
		amount := amountToken.GravityCoin()
		feeToken, err := types.NewInternalERC20Token(sdk.NewIntFromUint64(v), myTokenContractAddr)
		require.NoError(t, err)
		fee := feeToken.GravityCoin()
		_, err = input.GravityKeeper.AddToOutgoingPool(sdkCtx, mySender, *receiver, amount, fee)
		require.NoError(t, err)
		// Should create:
		// 1: amount 100, fee 2
		// 2: amount 101, fee 3
		// 3: amount 102, fee 2
		// 4: amount 104, fee 1
	}

	// when
	sdkCtx = sdkCtx.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	// Should contain 2 and 3 from above
	_, err = input.GravityKeeper.BuildOutgoingTXBatch(sdkCtx, *tokenContract, 2)
	require.NoError(t, err)

	// Should receive 1 and 4 unbatched, 2 and 3 batched in response
	response, err := k.GetPendingSendToEth(ctx, &types.QueryPendingSendToEth{SenderAddress: mySender.String()})
	require.NoError(t, err)
	expectedRes := types.QueryPendingSendToEthResponse{TransfersInBatches: []types.OutgoingTransferTx{
		{
			Id:          2,
			Sender:      "gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm",
			DestAddress: "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7",
			Erc20Token: types.ERC20Token{
				Contract: "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5",
				Amount:   sdk.NewInt(101),
			},
			Erc20Fee: types.ERC20Token{
				Contract: "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5",
				Amount:   sdk.NewInt(3),
			},
		},
		{
			Id:          3,
			Sender:      "gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm",
			DestAddress: "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7",
			Erc20Token: types.ERC20Token{
				Contract: "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5",
				Amount:   sdk.NewInt(102),
			},
			Erc20Fee: types.ERC20Token{
				Contract: "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5",
				Amount:   sdk.NewInt(2),
			},
		},
	},

		UnbatchedTransfers: []types.OutgoingTransferTx{
			{
				Id:          1,
				Sender:      "gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm",
				DestAddress: "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7",
				Erc20Token: types.ERC20Token{
					Contract: "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5",
					Amount:   sdk.NewInt(100),
				},
				Erc20Fee: types.ERC20Token{
					Contract: "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5",
					Amount:   sdk.NewInt(2),
				},
			},
			{
				Id:          4,
				Sender:      "gravity1ahx7f8wyertuus9r20284ej0asrs085ceqtfnm",
				DestAddress: "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7",
				Erc20Token: types.ERC20Token{
					Contract: "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5",
					Amount:   sdk.NewInt(103),
				},
				Erc20Fee: types.ERC20Token{
					Contract: "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5",
					Amount:   sdk.NewInt(1),
				},
			},
		},
	}

	assert.Equal(t, &expectedRes, response, "json is equal")
}
