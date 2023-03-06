package main

import (
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	palomaapp "github.com/palomachain/paloma/app"
)

type (
	CustomAppConfig struct {
		serverconfig.Config

		WASM WASMConfig `mapstructure:"wasm"`
	}

	// WASMConfig defines configuration for the wasm module.
	WASMConfig struct {
		// QueryGasLimit defines the maximum sdk gas (wasm and storage) that we allow
		// for any x/wasm "smart" queries.
		QueryGasLimit uint64 `mapstructure:"query_gas_limit"`

		// LRUSize defines the number of wasm VM instances we keep cached in memory
		// for speed-up.
		//
		// Warning: This is currently unstable and may lead to crashes, best to keep
		// for 0 unless testing locally.
		LRUSize uint64 `mapstructure:"lru_size"`
	}
)

// initAppConfig helps to override default appConfig template and configs.
// It will return <"", nil> if no custom configuration is required for the
// application.
func initAppConfig() (string, interface{}) {
	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := serverconfig.DefaultConfig()

	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	srvCfg.MinGasPrices = "0" + palomaapp.BondDenom

	customAppConfig := CustomAppConfig{
		Config: *srvCfg,
		WASM: WASMConfig{
			LRUSize:       1,
			QueryGasLimit: 300000,
		},
	}

	customAppTemplate := serverconfig.DefaultConfigTemplate + `
[wasm]
# This is the maximum sdk gas (wasm and storage) that we allow for any x/wasm "smart" queries
query_gas_limit = 300000
# This is the number of wasm vm instances we keep cached in memory for speed-up
# Warning: this is currently unstable and may lead to crashes, best to keep for 0 unless testing locally
lru_size = 0`

	return customAppTemplate, customAppConfig
}
