package main

import (
	"fmt"
	"os"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoprint"
	"github.com/katexochen/ttrpcurl"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "",
		Long:  "",
		RunE:  runDescribe,
	}

	return cmd
}

func runDescribe(cmd *cobra.Command, args []string) error {
	flags, err := parseDescribeFlags(cmd)
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

	printer := &protoprint.Printer{
		Compact:                  true,
		OmitComments:             protoprint.CommentsAll,
		SortElements:             true,
		ForceFullyQualifiedNames: true,
	}

	switch len(args) {
	case 0:
		for i := 0; i < fileDesc.Services().Len(); i++ {
			svc := fileDesc.Services().Get(i)
			fmt.Printf("%s is a service:\n", svc.FullName())

			wrappedSvc, err := desc.WrapDescriptor(svc)
			if err != nil {
				return fmt.Errorf("wrap descriptor: %w", err)
			}

			txt, err := printer.PrintProtoToString(wrappedSvc)
			if err != nil {
				return fmt.Errorf("print proto to string: %w", err)
			}

			fmt.Println(txt)
		}

		return nil
	case 1:
		serviceID, err := ttrpcurl.ServiceIdentifierFromFQN(args[0])
		if err != nil {
			return fmt.Errorf("parse service identifier: %w", err)
		}

		svc := fileDesc.Services().ByName(protoreflect.Name(serviceID.Service()))
		if svc == nil {
			return fmt.Errorf("service %q not found", args[0])
		}

		fmt.Printf("%s is a service:\n", svc.FullName())

		wrappedSvc, err := desc.WrapDescriptor(svc)
		if err != nil {
			return fmt.Errorf("wrap descriptor: %w", err)
		}

		txt, err := printer.PrintProtoToString(wrappedSvc)
		if err != nil {
			return fmt.Errorf("print proto to string: %w", err)
		}

		fmt.Println(txt)

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
