package main

import (
	"fmt"

	"github.com/fullstorydev/grpcurl"
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
		Args: cobra.MaximumNArgs(1),
		RunE: runDescribe,
	}

	cmd.Flags().Bool("msg-template", false, prettify(`
		When describing messages, show a template of input data.`))
	cmd.Flags().String("format", "json", prettify(`
		Format in which the message template should be printed. The allowed values
		are 'json' or 'text' for the protobuf text format.`))

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

	if len(args) == 0 {
		services := source.GetServices()
		for _, svc := range services {
			proroSnip, err := printer.PrintProtoToString(svc)
			if err != nil {
				return fmt.Errorf("printing proto to string: %w", err)
			}

			fmt.Printf("%s is a service:\n", svc.GetFullyQualifiedName())
			fmt.Printf("%s", proroSnip)
		}
		return nil
	}

	symbol, err := source.FindSymbol(args[0])
	if err != nil {
		return fmt.Errorf("finding symbol: %w", err)
	}

	symbolType, err := descriptorTypeStr(symbol)
	if err != nil {
		return fmt.Errorf("getting descriptor type: %w", err)
	}

	if flags.msgTemplate && symbolType != "message" {
		return fmt.Errorf("cannot show message template, %s is of type %s", args[0], symbolType)
	}

	proroSnip, err := printer.PrintProtoToString(symbol)
	if err != nil {
		return fmt.Errorf("printing proto to string: %w", err)
	}

	fmt.Printf("%s is a %s:\n", symbol.GetFullyQualifiedName(), symbolType)
	fmt.Printf("%s", proroSnip)

	if flags.msgTemplate {
		tmpl, err := createTemplate(symbol, source, flags.format)
		if err != nil {
			return fmt.Errorf("creating template: %w", err)
		}
		fmt.Println("\nMessage template:")
		fmt.Println(tmpl)
	}

	return nil
}

func createTemplate(symbol desc.Descriptor, source *proto.Source, format string) (string, error) {
	msg, ok := symbol.(*desc.MessageDescriptor)
	if !ok {
		return "", fmt.Errorf("symbol %s is not a message", symbol.GetFullyQualifiedName())
	}

	tmpl := grpcurl.MakeTemplate(msg)

	options := grpcurl.FormatOptions{EmitJSONDefaultFields: true}
	_, formatter, err := grpcurl.RequestParserAndFormatter(grpcurl.Format(format), source, nil, options)
	if err != nil {
		return "", fmt.Errorf("constructing formatter for %q: %w", format, err)
	}
	str, err := formatter(tmpl)
	if err != nil {
		return "", fmt.Errorf("printing template for message: %w", err)
	}

	return str, nil
}

type describeFlags struct {
	verbose     bool     // persistent
	proto       []string // persistent
	msgTemplate bool
	format      string
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
	f.msgTemplate, err = cmd.Flags().GetBool("msg-template")
	if err != nil {
		return nil, err
	}
	f.format, err = cmd.Flags().GetString("format")
	if err != nil {
		return nil, err
	}
	switch f.format {
	case "json":
	case "text":
	default:
		return nil, fmt.Errorf("unsupported format: %q", f.format)
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
