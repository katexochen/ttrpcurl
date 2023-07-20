package proto

import (
	"encoding/json"
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/desc/protoprint"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
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

func (s *Source) GetMessages() []*desc.MessageDescriptor {
	var messages []*desc.MessageDescriptor
	for _, fileDesc := range s.fileDescs {
		messages = append(messages, fileDesc.GetMessageTypes()...)
	}
	return messages
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

func (s *Source) FindMessage(message string) (*desc.MessageDescriptor, error) {
	symbol, err := s.FindSymbol(message)
	if err != nil {
		return nil, err
	}
	messageDesc, ok := symbol.(*desc.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("symbol %s is not a message", message)
	}
	return messageDesc, nil
}

func (s *Source) ListServices() ([]string, error) {
	panic("not implemented")
}

func (s *Source) AllExtensionsForType(_ string) ([]*desc.FieldDescriptor, error) {
	panic("not implemented")
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
			OmitComments:             protoprint.CommentsNonDoc,
			SortElements:             true,
			ForceFullyQualifiedNames: true,
		},
	}
}

func (p *Printer) PrintProtoToString(desc desc.Descriptor) (string, error) {
	return p.printer.PrintProtoToString(desc)
}

type Marshaler struct {
	Multiline bool
}

func (m Marshaler) Marshal(mes protoreflect.ProtoMessage) ([]byte, error) {
	b, err := protojson.Marshal(mes)
	if err != nil {
		return nil, fmt.Errorf("marshaling proto message: %w", err)
	}

	if m.Multiline {
		// The protojson package viciously adds random spaces between name and value
		// of JSON multiline output. As this is neither wanted for our users, nor in
		// the tests, we always use the protojson default marshaling and remarshal
		// with the standard library to get a clean output.
		//
		// See https://github.com/protocolbuffers/protobuf-go/blob/55f120eb3b91659cee86adeed925c825686556b0/internal/encoding/json/encode.go#L238-L243
		// for the gory details.

		var intermed any
		if err := json.Unmarshal(b, &intermed); err != nil {
			return nil, fmt.Errorf("unmarshaling json: %w", err)
		}

		b, err = json.MarshalIndent(intermed, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("remarshaling json: %w", err)
		}
	}

	return b, nil
}
