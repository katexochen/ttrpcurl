// Copyright (c) 2017 FullStory, Inc
// Based on https://github.com/fullstorydev/grpcurl/blob/9a59bed1d22aceb0719d4890dea642c058b0d623/internal/testing/test_server.go

package grpctest

import (
	"context"
	"io"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/containerd/ttrpc"
)

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-ttrpc_out=. --go-ttrpc_opt=paths=source_relative ./test.proto

// TestServer implements the TestService interface defined in example.proto.
type TestServer struct{}

// EmptyCall accepts one empty request and issues one empty response.
func (TestServer) EmptyCall(ctx context.Context, req *Empty) (*Empty, error) {
	_, failEarly, failLate := processMetadata(ctx)

	if failEarly != codes.OK {
		return nil, status.Error(failEarly, "fail")
	}
	if failLate != codes.OK {
		return nil, status.Error(failLate, "fail")
	}

	return req, nil
}

// UnaryCall accepts one request and issues one response. The response includes
// the client's payload as-is.
func (TestServer) UnaryCall(ctx context.Context, req *SimpleRequest) (*SimpleResponse, error) {
	_, failEarly, failLate := processMetadata(ctx)

	if failEarly != codes.OK {
		return nil, status.Error(failEarly, "fail")
	}

	if req.ResponseStatus != nil {
		return nil, status.Error(codes.Code(req.ResponseStatus.Code), req.ResponseStatus.Message)
	}

	if req.FillOauthScope {
		return nil, status.Error(codes.Unimplemented, "oauth scope not implemented")
	}

	resp := &SimpleResponse{}

	if req.FillUsername {
		resp.Username = "Paul"
	}

	if req.Payload != nil {
		resp.Payload = &Payload{
			Type: req.ResponseType,
			Body: make([]byte, req.ResponseSize),
		}
		for i := 0; i < int(req.ResponseSize); i++ {
			resp.Payload.Body[i] = byte('A')
		}
	}

	if failLate != codes.OK {
		return nil, status.Error(failLate, "fail")
	}

	return resp, nil
}

// StreamingOutputCall accepts one request and issues a sequence of responses
// (streamed download). The server returns the payload with client desired type
// and sizes as specified in the request's ResponseParameters.
func (TestServer) StreamingOutputCall(ctx context.Context, req *StreamingOutputCallRequest, str TestService_StreamingOutputCallServer) error {
	md, failEarly, failLate := processMetadata(ctx)
	ctx = ttrpc.WithMetadata(ctx, md)

	if failEarly != codes.OK {
		return status.Error(failEarly, "fail")
	}

	rsp := &StreamingOutputCallResponse{Payload: &Payload{}}
	for _, param := range req.ResponseParameters {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		delayMicros := int64(param.GetIntervalUs()) * int64(time.Microsecond)
		if delayMicros > 0 {
			time.Sleep(time.Duration(delayMicros))
		}
		sz := int(param.GetSize())
		buf := make([]byte, sz)
		for i := 0; i < sz; i++ {
			buf[i] = byte(i)
		}
		rsp.Payload.Type = req.ResponseType
		rsp.Payload.Body = buf
		if err := str.Send(rsp); err != nil {
			return err
		}
	}

	if failLate != codes.OK {
		return status.Error(failLate, "fail")
	}
	return nil
}

// StreamingInputCall accepts a sequence of requests and issues one response
// (streamed upload). The server returns the aggregated size of client payloads
// as the result.
func (TestServer) StreamingInputCall(ctx context.Context, str TestService_StreamingInputCallServer) (*StreamingInputCallResponse, error) {
	md, failEarly, failLate := processMetadata(ctx)
	ctx = ttrpc.WithMetadata(ctx, md)

	if failEarly != codes.OK {
		return nil, status.Error(failEarly, "fail")
	}

	sz := 0
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		req, err := str.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		sz += len(req.Payload.Body)
	}

	if failLate != codes.OK {
		return nil, status.Error(failLate, "fail")
	}

	return &StreamingInputCallResponse{AggregatedPayloadSize: int32(sz)}, nil
}

// FullDuplexCall accepts a sequence of requests with each request served by the
// server immediately. As one request could lead to multiple responses, this
// interface demonstrates the idea of full duplexing.
func (TestServer) FullDuplexCall(ctx context.Context, str TestService_FullDuplexCallServer) error {
	md, failEarly, failLate := processMetadata(ctx)
	ctx = ttrpc.WithMetadata(ctx, md)

	if failEarly != codes.OK {
		return status.Error(failEarly, "fail")
	}

	rsp := &StreamingOutputCallResponse{Payload: &Payload{}}
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		req, err := str.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		for _, param := range req.ResponseParameters {
			sz := int(param.GetSize())
			buf := make([]byte, sz)
			for i := 0; i < sz; i++ {
				buf[i] = byte(i)
			}
			rsp.Payload.Type = req.ResponseType
			rsp.Payload.Body = buf
			if err := str.Send(rsp); err != nil {
				return err
			}
		}
	}

	if failLate != codes.OK {
		return status.Error(failLate, "fail")
	}
	return nil
}

// HalfDuplexCall accepts a sequence of requests and issues a sequence of
// responses. The server buffers all the client requests and then serves them
// in order. A stream of responses is returned to the client once the client
// half-closes the stream.
func (TestServer) HalfDuplexCall(ctx context.Context, str TestService_HalfDuplexCallServer) error {
	md, failEarly, failLate := processMetadata(ctx)
	ctx = ttrpc.WithMetadata(ctx, md)

	if failEarly != codes.OK {
		return status.Error(failEarly, "fail")
	}

	var reqs []*StreamingOutputCallRequest
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		req, err := str.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		reqs = append(reqs, req)
	}
	rsp := &StreamingOutputCallResponse{}
	for _, req := range reqs {
		rsp.Payload = req.Payload
		if err := str.Send(rsp); err != nil {
			return err
		}
	}

	if failLate != codes.OK {
		return status.Error(failLate, "fail")
	}
	return nil
}

const (
	// MetadataReplyHeaders is a request header that contains values that will
	// be echoed back to the client as response headers. The format of the value
	// is "key: val". To have the server reply with more than one response
	// header, supply multiple values in request metadata.
	MetadataReplyHeaders = "reply-with-metadata"
	// MetadataFailEarly is a request header that, if present and not zero,
	// indicates that the RPC should fail immediately with that code.
	MetadataFailEarly = "fail-early"
	// MetadataFailLate is a request header that, if present and not zero,
	// indicates that the RPC should fail at the end with that code. This is
	// different from MetadataFailEarly only for streaming calls. An early
	// failure means the call to fail before any request stream is read or any
	// response stream is generated. A late failure means the entire request and
	// response streams will be consumed/processed and only then will the error
	// code be sent.
	MetadataFailLate = "fail-late"
)

func processMetadata(ctx context.Context) (replyMD ttrpc.MD, failEarly, failLate codes.Code) {
	md, ok := ttrpc.GetMetadata(ctx)
	if !ok {
		return nil, codes.OK, codes.OK
	}

	failEarly = toCode(md[MetadataFailEarly])
	failLate = toCode(md[MetadataFailLate])

	replyHeaders := md[MetadataReplyHeaders]
	replyMD = make(ttrpc.MD)
	replyMD.Set(MetadataReplyHeaders, replyHeaders...)

	return replyMD, failEarly, failLate
}

func toCode(vals []string) codes.Code {
	if len(vals) == 0 {
		return codes.OK
	}
	i, err := strconv.Atoi(vals[len(vals)-1])
	if err != nil {
		return codes.Code(i)
	}
	return codes.Code(i)
}

var _ TestServiceService = TestServer{}
