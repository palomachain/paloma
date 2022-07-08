package types

func (msg *ChainInfo) IsActive() bool {
	return msg.SmartContractAddr != "" && len(msg.SmartContractUniqueID) > 0
}

func (msg *SmartContract) GetUniqueID32() (res [32]byte) {
	copy(res[:], msg.UniqueID)
	return
}
