package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

var (
	version = "0.0.0-dev"
	commit  = "HEAD"
	date    = "unknown"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	rootCmd := newRootCmd()
	ctx, cancel := signalContext(context.Background(), os.Interrupt)
	defer cancel()
	return rootCmd.ExecuteContext(ctx)
}

func signalContext(ctx context.Context, sig os.Signal) (context.Context, context.CancelFunc) {
	sigCtx, stop := signal.NotifyContext(ctx, sig)
	done := make(chan struct{}, 1)
	stopDone := make(chan struct{}, 1)

	go func() {
		defer func() { stopDone <- struct{}{} }()
		defer stop()
		select {
		case <-sigCtx.Done():
			fmt.Println(" Signal caught. Shutting down. Press ctrl+c again to terminate the program immediately.")
		case <-done:
		}
	}()

	cancelFunc := func() {
		done <- struct{}{}
		<-stopDone
	}

	return sigCtx, cancelFunc
}

func preRunRoot(cmd *cobra.Command, _ []string) {
	cmd.SilenceUsage = true
}
