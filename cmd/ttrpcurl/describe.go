package main

import (
	"fmt"
	"os"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoprint"
	"github.com/katexochen/ttrpcurl"
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

	printer := &protoprint.Printer{
		Compact:                  true,
		OmitComments:             protoprint.CommentsAll,
		SortElements:             true,
		ForceFullyQualifiedNames: true,
	}

	file, err := desc.WrapFile(fileDesc)
	if err != nil {
		return fmt.Errorf("wrapping file descriptor: %w", err)
	}

	switch len(args) {
	case 0:
		services := file.GetServices()
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
		symbol := file.FindSymbol(args[0])
		if symbol == nil {
			return fmt.Errorf("symbol %q not found", args[0])
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
