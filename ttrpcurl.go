package ttrpcurl

import (
	"context"
	"fmt"
	"net"

	"github.com/containerd/ttrpc"
	"github.com/jhump/protoreflect/desc"
	"github.com/katexochen/ttrpcurl/proto"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/dynamicpb"
)

func Execute(protoFileNames []string, socket string, methodFQN string, reqBytes []byte) error {
	parser := proto.NewParser()
	source, err := parser.ParseFiles(protoFileNames...)
	if err != nil {
		return fmt.Errorf("parsing proto files: %w", err)
	}

	symbol, err := source.FindSymbol(methodFQN)
	if err != nil {
		return fmt.Errorf("finding symbol %q: %w", methodFQN, err)
	}

	mth, ok := symbol.(*desc.MethodDescriptor)
	if !ok {
		return fmt.Errorf("symbol %q is not a method", methodFQN)
	}

	req := dynamicpb.NewMessage(mth.GetInputType().UnwrapMessage())
	resp := dynamicpb.NewMessage(mth.GetOutputType().UnwrapMessage())

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

	err = client.Call(
		context.Background(),
		mth.GetService().GetFullyQualifiedName(),
		mth.GetName(),
		req,
		resp,
	)
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
