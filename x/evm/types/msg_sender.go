package types

import "context"

//go:generate mockery --name=MsgSender

type MsgSender interface {
	SendValsetMsgForChain(ctx context.Context, chainInfo *ChainInfo, valset Valset, assignee string) error
}
