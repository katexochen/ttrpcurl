package proto

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/desc/protoprint"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Source struct {
	fileDescs         []*desc.FileDescriptor
	includedFileDescs []*desc.FileDescriptor
}

func NewSource(files, includedFiles []*desc.FileDescriptor) *Source {
	return &Source{
		fileDescs:         files,
		includedFileDescs: includedFiles,
	}
}

func (s *Source) GetServices() []*desc.ServiceDescriptor {
	services := make(map[string]*desc.ServiceDescriptor, 0)

	// Included files first, so that they can be overridden.
	for _, fileDesc := range s.includedFileDescs {
		for _, service := range fileDesc.GetServices() {
			services[service.GetFullyQualifiedName()] = service
		}
	}
	for _, fileDesc := range s.fileDescs {
		for _, service := range fileDesc.GetServices() {
			services[service.GetFullyQualifiedName()] = service
		}
	}

	return mapToSortedSlice(services)
}

func (s *Source) GetMessages() []*desc.MessageDescriptor {
	messages := make(map[string]*desc.MessageDescriptor, 0)

	// Included files first, so that they can be overridden.
	for _, fileDesc := range s.includedFileDescs {
		for _, message := range fileDesc.GetMessageTypes() {
			messages[message.GetFullyQualifiedName()] = message
		}
	}
	for _, fileDesc := range s.fileDescs {
		for _, message := range fileDesc.GetMessageTypes() {
			messages[message.GetFullyQualifiedName()] = message
		}
	}

	return mapToSortedSlice(messages)
}

func (s *Source) FindSymbol(symbol string) (desc.Descriptor, error) {
	// User-defined symbols first, so the are chosen over built-in symbols.
	for _, fileDesc := range s.fileDescs {
		if symbol := fileDesc.FindSymbol(symbol); symbol != nil {
			return symbol, nil
		}
	}
	for _, fileDesc := range s.includedFileDescs {
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
	symbol, err := s.FindSymbol(service)
	if err != nil {
		return nil, err
	}
	serviceDesc, ok := symbol.(*desc.ServiceDescriptor)
	if !ok {
		return nil, fmt.Errorf("symbol %s is not a service", service)
	}
	return serviceDesc, nil
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

func (p *Parser) ParseFiles(filenames ...string) ([]*desc.FileDescriptor, error) {
	return p.parser.ParseFiles(filenames...)
}

func (p *Parser) WalkAndParse(fsys fs.FS, path string) ([]*desc.FileDescriptor, error) {
	entries, err := fs.ReadDir(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	var filenames []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filenames = append(filenames, entry.Name())
	}

	p.parser.Accessor = func(filename string) (io.ReadCloser, error) {
		return fsys.Open(filepath.Join(path, filename))
	}
	defer func() { p.parser.Accessor = nil }()

	return p.parser.ParseFiles(filenames...)
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

type fullyQualified interface {
	GetFullyQualifiedName() string
}

func mapToSortedSlice[T fullyQualified](m map[string]T) []T {
	s := make([]T, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}
	sort.Slice(s, func(i, j int) bool {
		return s[i].GetFullyQualifiedName() < s[j].GetFullyQualifiedName()
	})
	return s
}
