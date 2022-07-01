package types

func (msg *ChainInfo) IsActive() bool {
	return msg.GetStatus() == ChainInfo_ACTIVE && msg.SmartContractAddr != "" && msg.SmartContractID != ""
}
