package keeper

import (
	"math/big"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/palomachain/paloma/testutil/sample"
	"github.com/palomachain/paloma/x/evm/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vizualni/whoops"
)

func TestSubmittingNewMessageIsPuttingANewMessageToTheQueue(t *testing.T) {
	keeper, ms, ctx := newEvmKeeper(t)
	msgSvr := NewMsgServerImpl(*keeper)
	msg := &types.MsgSubmitNewJob{
		Creator:                 sdk.AccAddress("bla").String(),
		HexSmartContractAddress: common.BytesToAddress([]byte("abc")).String(),
		Abi:                     sample.SimpleABI,
		Method:                  "store",
		HexPayload: func() string {
			evm := whoops.Must(abi.JSON(strings.NewReader(sample.SimpleABI)))
			bz := whoops.Must(evm.Pack("store", big.NewInt(1337)))
			return common.Bytes2Hex(bz)
		}(),
	}

	ms.ConsensusKeeper.On("PutMessageForSigning", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	_, err := msgSvr.SubmitNewJob(sdk.WrapSDKContext(ctx), msg)

	require.NoError(t, err)
}
