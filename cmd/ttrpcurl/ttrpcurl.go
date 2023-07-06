package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/katexochen/ttrpcurl"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use:              "ttrpcurl [flags] <socket> <method>",
		Short:            "Make ttrpc calls based on a proto file",
		Version:          version,
		PersistentPreRun: preRunRoot,
		Args: cobra.MatchAll(
			cobra.ExactArgs(2),
		),
		RunE: runRootCmd,
	}

	rootCmd.SetOut(os.Stdout)

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	rootCmd.Flags().StringP("data", "d", "", "Data to send, @ means read from stdin")
	rootCmd.Flags().String("format", "json", "Format of data to send")
	// rootCmd.Flags().Bool("allow-unknown-fields", false, "Allow unknown fields")
	// rootCmd.Flags().String("connect-timeout", "10s", "Maximum time allowed for connection")
	rootCmd.Flags().String("proto", "", "Path to proto file")

	rootCmd.InitDefaultVersionFlag()
	rootCmd.SetVersionTemplate(
		fmt.Sprintf("ttrpcurl - Make ttrpc calls based on a proto file\n\nversion   %s\ncommit    %s\nbuilt at  %s\n", version, commit, date),
	)

	return rootCmd
}

func runRootCmd(cmd *cobra.Command, args []string) error {
	flags, err := parseFlags(cmd)
	if err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	fullMethName := strings.Split(args[1], ".")
	if len(fullMethName) != 3 {
		return fmt.Errorf("invalid method name: %s", args[1])
	}

	packageName := fullMethName[0]
	serviceName := fullMethName[1]
	methodName := fullMethName[2]

	var data []byte
	if flags.data == "@" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read data from stdin: %w", err)
		}
	} else {
		data, err = os.ReadFile(flags.data)
		if err != nil {
			return fmt.Errorf("read data file: %w", err)
		}
	}

	return ttrpcurl.Execute(
		flags.proto,
		args[0],
		packageName,
		serviceName,
		methodName,
		data,
	)
}

type flags struct {
	verbose bool
	data    string
	format  string
	proto   string
}

func parseFlags(cmd *cobra.Command) (*flags, error) {
	f := &flags{}

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
