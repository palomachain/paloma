package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type MsgSender interface {
	SendValsetMsgForChain(ctx sdk.Context, chainInfo *ChainInfo, valset Valset) error
}
