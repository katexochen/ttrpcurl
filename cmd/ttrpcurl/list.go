package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/jhump/protoreflect/desc"
	"github.com/katexochen/ttrpcurl"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [flags] [symbol]",
		Example: "ttrpcurl list --proto=api.proto package.Service",
		Short:   "List available protobuf services or methods of a service",
		Long: prettify(`
			Show a list of available services or methods, based on the given proto source.
			If the symbol is a fully-qualified name of a protobuf service, formatted
			like '[package.]service' or '[package/]service', the methods of that service
			are listed. If no symbol is given, all available services are listed.`),
		RunE: runList,
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
		return fmt.Errorf("opening proto file: %w", err)
	}
	defer f.Close()

	parser := ttrpcurl.NewProtoParser()

	fileDesc, err := parser.ParseFile(flags.proto[0], f)
	if err != nil {
		return fmt.Errorf("parsing proto file: %w", err)
	}

	file, err := desc.WrapFile(fileDesc)
	if err != nil {
		return fmt.Errorf("wrapping file descriptor: %w", err)
	}

	switch len(args) {
	case 0:
		var svcNames []string
		for _, svc := range file.GetServices() {
			svcNames = append(svcNames, svc.GetFullyQualifiedName())
		}
		sort.Strings(svcNames)
		for _, svcName := range svcNames {
			fmt.Println(svcName)
		}
		return nil
	case 1:
		svc := file.FindService(args[0])
		if svc == nil {
			return fmt.Errorf("service %q not found", args[0])
		}
		var methodNames []string
		for _, method := range svc.GetMethods() {
			methodNames = append(methodNames, method.GetFullyQualifiedName())
		}
		sort.Strings(methodNames)
		for _, methodName := range methodNames {
			fmt.Println(methodName)
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
