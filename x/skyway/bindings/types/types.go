package types

type ERC20ToDenom struct {
	Erc20            string `json:"erc20,omitempty"`
	Denom            string `json:"denom,omitempty"`
	ChainReferenceId string `json:"chainReferenceId,omitempty"`
}
