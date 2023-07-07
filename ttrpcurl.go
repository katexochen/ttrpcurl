package ttrpcurl

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/containerd/ttrpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type MethodIdentifier struct {
	Package string
	Service string
	Method  string
}

func MethodIdentifierFromFQN(fqn string) (MethodIdentifier, error) {
	parts := strings.Split(fqn, ".")
	switch len(parts) {
	case 3:
		return MethodIdentifier{
			Package: parts[0],
			Service: parts[1],
			Method:  parts[2],
		}, nil
	case 2:
		return MethodIdentifier{
			Service: parts[0],
			Method:  parts[1],
		}, nil
	default:
		return MethodIdentifier{}, fmt.Errorf("failed to parse method identifier: %q isn't a fully qualified name", fqn)
	}
}

type ProtoParser struct {
	reportHandler *reporter.Handler
}

func NewProtoParser() *ProtoParser {
	repHandler := reporter.NewHandler(
		reporter.NewReporter(
			// TODO: look for upstream example of how to use this
			func(err reporter.ErrorWithPos) error { fmt.Printf(err.Error()); return err },
			func(ewp reporter.ErrorWithPos) { fmt.Printf(ewp.Error()) },
		),
	)

	return &ProtoParser{reportHandler: repHandler}
}

func (p *ProtoParser) ParseFile(filename string, r io.Reader) (*descriptorpb.FileDescriptorProto, error) {
	ast, err := parser.Parse(filename, r, p.reportHandler)
	if err != nil {
		return nil, fmt.Errorf("parsing proto file %q: %w", filename, err)
	}

	validateResult := true
	result, err := parser.ResultFromAST(ast, validateResult, p.reportHandler)
	if err != nil {
		return nil, fmt.Errorf("getting result from AST: %w", err)
	}

	return result.FileDescriptorProto(), nil
}

func Execute(filename string, socket string, methodFQN string, reqBytes []byte) error {
	methodID, err := MethodIdentifierFromFQN(methodFQN)
	if err != nil {
		return err
	}

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	parser := NewProtoParser()
	fileDescProto, err := parser.ParseFile(filename, f)
	if err != nil {
		return err
	}

	fileDesc, err := protodesc.NewFile(fileDescProto, nil)
	if err != nil {
		return err
	}
	svc := fileDesc.Services().ByName(protoreflect.Name(methodID.Service))
	if svc == nil {
		return fmt.Errorf("service with name " + methodID.Service + " not found")
	}

	mth := svc.Methods().ByName(protoreflect.Name(methodID.Method))
	if mth == nil {
		return fmt.Errorf("method with name " + methodID.Method + " not found")
	}

	req := dynamicpb.NewMessage(mth.Input())
	resp := dynamicpb.NewMessage(mth.Output())

	if err := protojson.Unmarshal(reqBytes, req); err != nil {
		return err
	}

	if !req.IsValid() {
		return fmt.Errorf("invalid request")
	}

	con, err := net.Dial("unix", socket)
	if err != nil {
		return err
	}

	client := ttrpc.NewClient(con)
	defer client.Close()

	err = client.Call(context.Background(), methodID.Package+"."+methodID.Service, methodID.Method, req, resp)
	if err != nil {
		return err
	}

	return nil
}

type Client struct {
	ttrpc ttrpcClient
}

func NewClient(conn net.Conn) *Client {
	return &Client{ttrpc: ttrpc.NewClient(conn)}
}

type ttrpcClient interface {
	Call(ctx context.Context, service, method string, req, resp interface{}) error
}
