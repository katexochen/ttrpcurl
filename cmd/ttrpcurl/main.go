package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

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

	rootCmd.PersistentPreRun = preRunRoot
	rootCmd.SetOut(os.Stdout)

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	rootCmd.AddCommand(
		newListCommand(),
		newDescribeCommand(),
	)

	rootCmd.Version = version
	rootCmd.InitDefaultVersionFlag()
	rootCmd.SetVersionTemplate(
		fmt.Sprintf("ttrpcurl - Make ttrpc calls based on a proto file\n\nversion   %s\ncommit    %s\nbuilt at  %s\n", version, commit, date),
	)

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

func prettify(docString string) string {
	parts := strings.Split(docString, "\n")

	// cull empty lines and also remove trailing and leading spaces
	// from each line in the doc string
	j := 0
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		parts[j] = part
		j++
	}

	return strings.Join(parts[:j], "\n")
}
