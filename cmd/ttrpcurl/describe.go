package main

import (
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/katexochen/ttrpcurl/proto"
	"github.com/spf13/cobra"
)

func newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "describe [flags] [symbol]",
		Example: "ttrpcurl describe --proto=api.proto package.Service.Method",
		Short:   "Describe a protobuf symbol from the given source",
		Long: prettify(`
			Show the descriptor of a given symbol based on the given proto source.
			The symbol should be a fully-qualified name of a protobuf message,
			enum, service, method, or field. If no symbol is given the descriptors
			for all exposed or known services are shown`),
		RunE: runDescribe,
	}

	return cmd
}

func runDescribe(cmd *cobra.Command, args []string) error {
	flags, err := parseDescribeFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	parser := proto.NewParser()
	source, err := parser.ParseFiles(flags.proto...)
	if err != nil {
		return fmt.Errorf("parsing proto files: %w", err)
	}

	printer := proto.NewPrinter()

	switch len(args) {
	case 0:
		services := source.GetServices()
		for _, svc := range services {
			proroSnip, err := printer.PrintProtoToString(svc)
			if err != nil {
				return fmt.Errorf("printing proto to string: %w", err)
			}

			fmt.Printf("%s is a service:\n", svc.GetFullyQualifiedName())
			fmt.Printf(proroSnip)
		}
		return nil
	case 1:
		symbol, err := source.FindSymbol(args[0])
		if err != nil {
			return fmt.Errorf("finding symbol: %w", err)
		}

		symbolType, err := descriptorTypeStr(symbol)
		if err != nil {
			return fmt.Errorf("getting descriptor type: %w", err)
		}

		proroSnip, err := printer.PrintProtoToString(symbol)
		if err != nil {
			return fmt.Errorf("printing proto to string: %w", err)
		}

		fmt.Printf("%s is a %s:\n", symbol.GetFullyQualifiedName(), symbolType)
		fmt.Printf(proroSnip)
		return nil
	default:
		return fmt.Errorf("too many arguments")
	}
}

type describeFlags struct {
	verbose bool     // persistent
	proto   []string // persistent
}

func parseDescribeFlags(cmd *cobra.Command) (*describeFlags, error) {
	f := &describeFlags{}

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

func descriptorTypeStr(d desc.Descriptor) (string, error) {
	switch d.(type) {
	case *desc.MessageDescriptor:
		return "message", nil
	case *desc.EnumDescriptor:
		return "enum", nil
	case *desc.ServiceDescriptor:
		return "service", nil
	case *desc.MethodDescriptor:
		return "method", nil
	case *desc.FieldDescriptor:
		return "field", nil
	default:
		return "", fmt.Errorf("unknown descriptor type: %T", d)
	}
}
