package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/containerd/ttrpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func main() {
	// importPaths := []string{"."}
	filenames := []string{"getresource.proto"}
	serviceName := "GetResourceService"
	methodName := "GetResource"

	f, err := os.Open(filenames[0])
	if err != nil {
		panic(err)
	}

	repHandler := reporter.NewHandler(
		reporter.NewReporter(
			func(err reporter.ErrorWithPos) error { fmt.Printf(err.Error()); return err },
			func(ewp reporter.ErrorWithPos) { fmt.Printf(ewp.Error()) },
		),
	)
	ast, err := parser.Parse(filenames[0], f, repHandler)
	if err != nil {
		panic(err)
	}

	result, err := parser.ResultFromAST(ast, true, repHandler)
	if err != nil {
		panic(err)
	}

	fileDescProto := result.FileDescriptorProto()

	fileDesc, err := protodesc.NewFile(fileDescProto, nil)
	if err != nil {
		panic(err)
	}
	svc := fileDesc.Services().ByName(protoreflect.Name(serviceName))
	if svc == nil {
		panic("service with name " + serviceName + " not found")
	}
	fmt.Println(svc.FullName())

	mth := svc.Methods().ByName(protoreflect.Name(methodName))
	if mth == nil {
		panic("method with name " + methodName + " not found")
	}
	fmt.Println(mth.FullName())

	req := dynamicpb.NewMessage(mth.Input())
	resp := dynamicpb.NewMessage(mth.Output())

	reqBytes := []byte(`{"KbcName": "name","KbsUri":"uri","ResourcePath":"path"}`)

	if err := protojson.Unmarshal(reqBytes, req); err != nil {
		panic(err)
	}

	fmt.Println(req)
	if !req.IsValid() {
		panic("invalid request")
	}

	con, err := net.Dial("unix", "./ttrpc-test.sock")
	if err != nil {
		panic(err)
	}

	client := ttrpc.NewClient(con)
	defer client.Close()

	err = client.Call(context.Background(), serviceName, methodName, req, resp)
	if err != nil {
		panic(err)
	}
}
