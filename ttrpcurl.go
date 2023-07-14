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

type Client struct {
	ttrpc  ttrpcClient
	source *proto.Source
}

func NewClient(conn net.Conn, source *proto.Source) *Client {
	return &Client{
		ttrpc:  ttrpc.NewClient(conn),
		source: source,
	}
}

func (c *Client) Call(ctx context.Context, method string, reqBytes []byte) error {
	mth, err := c.source.FindMethod(method)
	if err != nil {
		return err
	}

	switch {
	case mth.IsClientStreaming() && mth.IsServerStreaming():
		return c.callBidirectionalSteaming(ctx, mth, reqBytes)
	case mth.IsClientStreaming():
		return c.callClientSteaming(ctx, mth, reqBytes)
	case mth.IsServerStreaming():
		return c.callServerSteaming(ctx, mth, reqBytes)
	default:
		return c.callUnary(ctx, mth, reqBytes)
	}
}

func (c *Client) callUnary(ctx context.Context, mth *desc.MethodDescriptor, reqBytes []byte) error {
	req := dynamicpb.NewMessage(mth.GetInputType().UnwrapMessage())
	resp := dynamicpb.NewMessage(mth.GetOutputType().UnwrapMessage())

	if err := protojson.Unmarshal(reqBytes, req); err != nil {
		return err
	}

	if !req.IsValid() {
		return fmt.Errorf("invalid request")
	}

	if err := c.ttrpc.Call(
		ctx,
		mth.GetService().GetFullyQualifiedName(),
		mth.GetName(),
		req,
		resp,
	); err != nil {
		return err
	}

	return nil
}

func (c *Client) callServerSteaming(ctx context.Context, mth *desc.MethodDescriptor, reqBytes []byte) error {
	panic("server streaming not implemented")
}

func (c *Client) callClientSteaming(ctx context.Context, mth *desc.MethodDescriptor, reqBytes []byte) error {
	panic("client streaming not implemented")
}

func (c *Client) callBidirectionalSteaming(ctx context.Context, mth *desc.MethodDescriptor, reqBytes []byte) error {
	panic("bidirectional streaming not implemented")
}

type ttrpcClient interface {
	Call(ctx context.Context, service, method string, req, resp interface{}) error
}
