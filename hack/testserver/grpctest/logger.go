// Copyright (c) 2017 FullStory, Inc
// Based on https://github.com/fullstorydev/grpcurl/blob/9a59bed1d22aceb0719d4890dea642c058b0d623/internal/testing/cmd/testserver/testserver.go

package grpctest

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/containerd/ttrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var id int32

func unaryLogger(ctx context.Context, unmarshaler ttrpc.Unmarshaler, info *ttrpc.UnaryServerInfo, method ttrpc.Method) (any, error) {
	i := atomic.AddInt32(&id, 1) - 1
	log.Printf("<%d> start: %s\n", i, info.FullMethod)
	start := time.Now()
	rsp, err := method(ctx, unmarshaler)
	var code codes.Code
	if stat, ok := status.FromError(err); ok {
		code = stat.Code()
	} else {
		code = codes.Unknown
	}
	log.Printf("<%d> completed: %v (%d) %v\n", i, code, code, time.Since(start))
	return rsp, err
}

func streamLogger(ctx context.Context, ss ttrpc.StreamServer, info *ttrpc.StreamServerInfo, handler ttrpc.StreamHandler) (any, error) {
	i := atomic.AddInt32(&id, 1) - 1
	start := time.Now()
	log.Printf("<%d> start: %s\n", i, info.FullMethod)
	resp, err := handler(ctx, loggingStream{ss: ss, id: i})
	var code codes.Code
	if stat, ok := status.FromError(err); ok {
		code = stat.Code()
	} else {
		code = codes.Unknown
	}
	log.Printf("<%d> completed: %v(%d) %v\n", i, code, code, time.Since(start))
	return resp, err
}

type loggingStream struct {
	ss ttrpc.StreamServer
	id int32
}

func (l loggingStream) SendMsg(m interface{}) error {
	err := l.ss.SendMsg(m)
	if err == nil {
		log.Printf("stream <%d>: sent message\n", l.id)
	}
	return err
}

func (l loggingStream) RecvMsg(m interface{}) error {
	err := l.ss.RecvMsg(m)
	if err == nil {
		log.Printf("stream <%d>: received message\n", l.id)
	}
	return err
}
