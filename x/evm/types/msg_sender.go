package types

import "context"

type MsgSender interface {
	SendValsetMsgForChain(ctx context.Context, chainInfo *ChainInfo, valset Valset, assignee string) error
}
