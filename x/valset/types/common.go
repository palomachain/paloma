package types

func (m MsgMetadata) GetSigners() []string {
	return m.Signers
}

func (m MsgMetadata) GetCreator() string {
	return m.Creator
}
