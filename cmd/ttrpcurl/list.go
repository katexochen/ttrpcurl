package main

import (
	"fmt"
	"os"

	"github.com/katexochen/ttrpcurl"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "",
		Long:  "",
		RunE:  runList,
	}

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	flags, err := parseListFlags(cmd)
	if err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	f, err := os.Open(flags.proto[0])
	if err != nil {
		return fmt.Errorf("open proto file: %w", err)
	}
	defer f.Close()

	parser := ttrpcurl.NewProtoParser()

	fileDesc, err := parser.ParseFile(flags.proto[0], f)
	if err != nil {
		return fmt.Errorf("parse proto file: %w", err)
	}

	switch len(args) {
	case 0:
		for i := 0; i < fileDesc.Services().Len(); i++ {
			fmt.Println(fileDesc.Services().Get(i).FullName())
		}

		return nil
	case 1:
		serviceID, err := ttrpcurl.ServiceIdentifierFromFQN(args[0])
		if err != nil {
			return fmt.Errorf("parse service identifier: %w", err)
		}

		serviceDesc := fileDesc.Services().ByName(protoreflect.Name(serviceID.Service()))
		if serviceDesc == nil {
			return fmt.Errorf("service %q not found", args[0])
		}

		for i := 0; i < serviceDesc.Methods().Len(); i++ {
			fmt.Println(serviceDesc.Methods().Get(i).FullName())
		}

		return nil
	default:
		return fmt.Errorf("too many arguments")
	}
}

type listFlags struct {
	verbose bool     // persistent
	proto   []string // persistent
}

func parseListFlags(cmd *cobra.Command) (*listFlags, error) {
	f := &listFlags{}

	var err error
	f.verbose, err = cmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, err
	}
	f.proto, err = cmd.Flags().GetStringSlice("proto")
	if err != nil {
		return nil, err
	}

	return f, nil
}
