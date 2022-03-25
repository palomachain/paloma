package app

import (
	"context"
	"os"
	"os/signal"
	"time"
)

// A runner is a term used in the jupiter chain to referr to the go-routines
// running alongside main chain.

type runnerer interface {
	Run(context.Context)
}

type runnerStarter func(ctx context.Context)

func runRunners(rootCtx context.Context, runners ...runnerer) {
	ctx, closeCtxFnc := context.WithCancel(rootCtx)

	quit := make(chan os.Signal, 5)

	signal.Notify(quit, os.Interrupt)

	for _, runner := range runners {
		go runner.Run(ctx)
	}

	// wait for the signal to be sent
	<-quit

	closeCtxFnc()

	// Giving some time for runners to clear their work and exit.
	time.Sleep(5 * time.Second)
}
