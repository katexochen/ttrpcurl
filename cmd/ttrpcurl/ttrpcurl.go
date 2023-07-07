package main

import (
	"fmt"
	"io"
	"os"

	"github.com/katexochen/ttrpcurl"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use:   "ttrpcurl [flags] <socket> <method>",
		Short: "Make ttrpc calls based on a proto file",
		Args: cobra.MatchAll(
			cobra.ExactArgs(2),
		),
		RunE: runRootCmd,
	}

	rootCmd.Flags().StringP("data", "d", "", "Data to send, @ means read from stdin")
	rootCmd.Flags().String("format", "json", "Format of data to send")
	// rootCmd.Flags().Bool("allow-unknown-fields", false, "Allow unknown fields")
	// rootCmd.Flags().String("connect-timeout", "10s", "Maximum time allowed for connection")
	rootCmd.Flags().String("proto", "", "Path to proto file")

	// Unused flags for compability with grpcurl
	rootCmd.Flags().Bool("plaintext", false, "")
	rootCmd.Flags().MarkHidden("plaintext")

	return rootCmd
}

func runRootCmd(cmd *cobra.Command, args []string) error {
	flags, err := parseRootFlags(cmd)
	if err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	if err := warnRootCompabilityFlags(cmd); err != nil {
		return fmt.Errorf("parsing compability flags: %w", err)
	}

	var data []byte
	if flags.data == "@" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read data from stdin: %w", err)
		}
	} else {
		data = []byte(flags.data)
	}

	return ttrpcurl.Execute(
		flags.proto,
		args[0],
		args[1],
		data,
	)
}

type rootFlags struct {
	verbose bool
	data    string
	format  string
	proto   string
}

func parseRootFlags(cmd *cobra.Command) (*rootFlags, error) {
	f := &rootFlags{}

	var err error
	f.verbose, err = cmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, err
	}
	f.data, err = cmd.Flags().GetString("data")
	if err != nil {
		return nil, err
	}
	f.format, err = cmd.Flags().GetString("format")
	if err != nil {
		return nil, err
	}
	f.proto, err = cmd.Flags().GetString("proto")
	if err != nil {
		return nil, err
	}

	return f, nil
}

func warnRootCompabilityFlags(cmd *cobra.Command) error {
	boolFlags := []struct {
		name    string
		warning string
	}{
		{"plaintext", "The flag deactivates TLS in grpcurl, but ttrpcurl communicates over a unix domain socket. It doesn't use TCP, so TLS isn't involved per default."},
	}

	for _, flag := range boolFlags {
		val, err := cmd.Flags().GetBool(flag.name)
		if err != nil {
			return err
		}
		if val {
			fmt.Printf("WARN: flag %s is unused and only provided for compability with grpcurl. %s\n", flag.name, flag.warning)
		}
	}

	return nil
}
