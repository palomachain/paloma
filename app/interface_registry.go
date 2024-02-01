package app

import (
	"fmt"

	"cosmossdk.io/x/tx/signing"

	"cosmossdk.io/core/address"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	params2 "github.com/palomachain/paloma/app/params"
	golangproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var InterfaceRegistry types.InterfaceRegistry

func init() {
	var err error

	if InterfaceRegistry, err = NewInterfaceRegistry(
		params2.AccountAddressPrefix,
		params2.ValidatorAddressPrefix,
	); err != nil {
		panic(err)
	}
}

//getCustomMsgSignerFn returns CustomSigners
func getCustomMsgSignerFn(a address.Codec, path []string) func(msg golangproto.Message) ([][]byte, error) {
	if len(path) == 0 {
		panic("path is expected to contain at least one value.")
	}
	return func(msg golangproto.Message) ([][]byte, error) {
		m := msg.ProtoReflect()
		fieldDesc := m.Descriptor().Fields().ByName(protoreflect.Name(path[len(path)-1]))
		if fieldDesc.Kind() != protoreflect.MessageKind {
			return nil, fmt.Errorf(
				"Expected for final field %s to be String type in path %+v for msg %+v.",
				path[len(path)-1],
				path,
				msg,
			)
		}
		signersDesc := fieldDesc.Message().Fields().ByName("signers")
		var res [][]byte
		if signersDesc.IsList() {
			signers := m.Get(fieldDesc).Message().Get(signersDesc).List()
			lenOfSigners := signers.Len()
			for j := 0; j < lenOfSigners; j++ {
				sgnr := signers.Get(j).String()
				sgnrBz, err := a.StringToBytes(sgnr)
				if err != nil {
					return nil, err
				}
				res = append(res, sgnrBz)
			}
		}

		return res, nil
	}
}

// NewInterfaceRegistry returns InterfaceRegistry with the customGetSigners
func NewInterfaceRegistry(addrPrefix string, valAddrPrefix string) (types.InterfaceRegistry, error) {
	messageArray := []protoreflect.FullName{
		"palomachain.paloma.valset.MsgAddExternalChainInfoForValidator",
		"palomachain.paloma.valset.MsgKeepAlive",
		"palomachain.paloma.evm.MsgDeployNewSmartContractRequest",
		"palomachain.paloma.evm.MsgRemoveSmartContractDeploymentRequest",
		"palomachain.paloma.consensus.MsgAddMessagesSignatures",
		"palomachain.paloma.consensus.MsgDeleteJob",
		"palomachain.paloma.consensus.MsgAddEvidence",
		"palomachain.paloma.consensus.MsgSetPublicAccessData",
		"palomachain.paloma.consensus.MsgSetErrorData",
		"palomachain.paloma.gravity.MsgSendToEth",
		"palomachain.paloma.gravity.MsgConfirmBatch",
		"palomachain.paloma.gravity.MsgSendToPalomaClaim",
		"palomachain.paloma.gravity.MsgBatchSendToEthClaim",
		"palomachain.paloma.gravity.MsgCancelSendToEth",
		"palomachain.paloma.gravity.MsgSubmitBadSignatureEvidence",
		"palomachain.paloma.gravity.MsgUpdateParams",
		"palomachain.paloma.paloma.MsgAddStatusUpdate",
		"palomachain.paloma.scheduler.MsgCreateJob",
		"palomachain.paloma.scheduler.MsgExecuteJob",
	}
	messages := make(map[protoreflect.FullName]signing.GetSignersFunc)
	for _, v := range messageArray {
		messages[v] = getCustomMsgSignerFn(addresscodec.NewBech32Codec(addrPrefix), []string{"metadata"})
	}

	return types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: gogoproto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          addresscodec.NewBech32Codec(addrPrefix),
			ValidatorAddressCodec: addresscodec.NewBech32Codec(valAddrPrefix),
			CustomGetSigners:      map[protoreflect.FullName]signing.GetSignersFunc(messages),
		},
	})
}
