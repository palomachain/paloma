package runner

func Start() {
	panic("main loop goes here")
	// var c terra.Client
	// modules := append(lens.ModuleBasics[:], byop.NewModule(
	// 	"testing",
	// 	(*types.MsgExecuteContract)(nil),
	// ))

	// modules := append(lens.ModuleBasics[:])
	// lensc, err := chain.NewChainClient(
	// 	&lens.ChainClientConfig{
	// 		Key:            "matija",
	// 		ChainID:        "columbus-5",
	// 		RPCAddr:        "http://127.0.0.1:22222",
	// 		KeyringBackend: "file",
	// 		KeyDirectory:   "/home/vizualni/.terra/",
	// 		AccountPrefix:  "terra",
	// 		Modules:        modules,
	// 		Debug:          true,
	// 		GasAdjustment:  1.1,
	// 		GasPrices:      "0.2056735uusd,0luna",
	// 		SignModeStr:    "direct",
	// 	},
	// 	"doesn-tmatter",
	// 	os.Stdin,
	// 	os.Stdout,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// c.LensClient = lensc
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()
	// c.ExecuteSmartContract(ctx)
}
