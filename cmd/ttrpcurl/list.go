package main

import (
	"fmt"
	"sort"

	"github.com/katexochen/ttrpcurl/proto"
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

	parser := proto.NewParser()
	source, err := parser.ParseFiles(flags.proto...)
	if err != nil {
		return fmt.Errorf("parsing proto files: %w", err)
	}

	switch len(args) {
	case 0:
		var svcNames []string
		for _, svc := range source.GetServices() {
			svcNames = append(svcNames, svc.GetFullyQualifiedName())
		}
		sort.Strings(svcNames)
		for _, svcName := range svcNames {
			fmt.Println(svcName)
		}
		return nil
	case 1:
		svc, err := source.FindService(args[0])
		if err != nil {
			return fmt.Errorf("finding service: %w", err)
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
