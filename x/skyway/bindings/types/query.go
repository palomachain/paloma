package types

type Query struct {
	Erc20ToDenoms *Erc20ToDenoms `json:"erc20_to_denoms,omitempty"`
}

type Erc20ToDenoms struct{}

type Erc20ToDenomsResponse struct {
	Denoms []ERC20ToDenom `json:"denoms"`
}
