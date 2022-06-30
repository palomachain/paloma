package types

func (msg *ChainInfo) IsActive() bool {
	return msg.SmartContractAddr != "" && msg.SmartContractID != ""
}
