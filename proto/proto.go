package proto

import (
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/desc/protoprint"
)

type Source struct {
	fileDescs []*desc.FileDescriptor
}

func (s *Source) GetServices() []*desc.ServiceDescriptor {
	var services []*desc.ServiceDescriptor
	for _, fileDesc := range s.fileDescs {
		services = append(services, fileDesc.GetServices()...)
	}
	return services
}

func (s *Source) FindSymbol(symbol string) (desc.Descriptor, error) {
	for _, fileDesc := range s.fileDescs {
		if symbol := fileDesc.FindSymbol(symbol); symbol != nil {
			return symbol, nil
		}
	}
	return nil, fmt.Errorf("symbol %s not found", symbol)
}

func (s *Source) FindMethod(method string) (*desc.MethodDescriptor, error) {
	symbol, err := s.FindSymbol(method)
	if err != nil {
		return nil, err
	}
	methodDesc, ok := symbol.(*desc.MethodDescriptor)
	if !ok {
		return nil, fmt.Errorf("symbol %s is not a method", method)
	}
	return methodDesc, nil
}

func (s *Source) FindService(service string) (*desc.ServiceDescriptor, error) {
	for _, fileDesc := range s.fileDescs {
		if service := fileDesc.FindService(service); service != nil {
			return service, nil
		}
	}
	return nil, fmt.Errorf("service %s not found", service)
}

type Parser struct {
	parser protoparse.Parser
}

func NewParser() *Parser {
	return &Parser{
		parser: protoparse.Parser{
			// ImportPaths:           importPaths,
			// InferImportPaths:      len(importPaths) == 0,
			IncludeSourceCodeInfo: true,
		},
	}
}

func (p *Parser) ParseFiles(filenames ...string) (*Source, error) {
	desc, err := p.parser.ParseFiles(filenames...)
	if err != nil {
		return nil, err
	}
	return &Source{fileDescs: desc}, nil
}

type Printer struct {
	printer protoprint.Printer
}

func NewPrinter() *Printer {
	return &Printer{
		printer: protoprint.Printer{
			Compact:                  true,
			OmitComments:             protoprint.CommentsAll,
			SortElements:             true,
			ForceFullyQualifiedNames: true,
		},
	}
}

func (p *Printer) PrintProtoToString(desc desc.Descriptor) (string, error) {
	return p.printer.PrintProtoToString(desc)
}
