package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetupPalomaPrefixes sets the Prefix for validator and Account
func SetupPalomaPrefixes() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("paloma", "pub")
	config.SetBech32PrefixForValidator("palomavaloper", "valoperpub")
}
