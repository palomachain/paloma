package terra

import (
	"context"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	chain "github.com/volumefi/cronchain/runner/client"
	"github.com/volumefi/cronchain/runner/client/terra/types"
)

type SmartContractExecution struct {
	Contract string
	Sender   string
	Payload  json.RawMessage
	Coins    sdk.Coins
}

type Client struct {
	LensClient *chain.LensClient
}

// TODO: this is currently oversimplified. Once we start using this for real we will adapt.
// Maybe better thing would be to actually use the "Invoke" method along with the grpc client.
func (c Client) ExecuteSmartContract(
	ctx context.Context,
	msg SmartContractExecution,
) (*sdk.TxResponse, error) {
	// TODO: do validations
	var msgExec types.MsgExecuteContract
	msgExec.Contract = msg.Contract
	msgExec.Sender = msg.Sender
	msgExec.Coins = msg.Coins
	msgExec.ExecuteMsg = msg.Payload
	return c.LensClient.SendMsg(ctx, &msgExec)
}
