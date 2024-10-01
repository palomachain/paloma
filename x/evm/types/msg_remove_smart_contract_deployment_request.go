package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/palomachain/paloma/v2/util/libmeta"
)

func (msg *MsgRemoveSmartContractDeploymentRequest) ValidateBasic() error {
	return libmeta.ValidateBasic(msg)
}

func (msg *MsgRemoveSmartContractDeploymentRequest) GetSigners() []sdk.AccAddress {
	return libmeta.GetSigners(msg)
}
