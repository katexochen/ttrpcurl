package main

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/containerd/ttrpc"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func main() {
	importPaths := []string{"."}
	filenames := []string{"getresource.proto"}
	serviceName := "getresource.GetResourceService"
	methodName := "GetResource"

	filenames, err := protoparse.ResolveFilenames(importPaths, filenames...)
	if err != nil {
		panic(err)
	}

	protoParser := protoparse.Parser{
		ImportPaths:           importPaths,
		InferImportPaths:      len(importPaths) == 0,
		IncludeSourceCodeInfo: true,
	}

	fds, err := protoParser.ParseFiles(filenames...)
	if err != nil {
		panic(err)
	}

	ds := make([]protoreflect.Descriptor, len(fds))
	for i, fd := range fds {
		ds[i] = fd.Unwrap()
	}

	for _, d := range ds {
		fmt.Println(d.FullName())
	}

	dsource, err := DescriptorSourceFromProtoFiles(importPaths, filenames...)
	if err != nil {
		panic(err)
	}

	fmt.Println(ListMethods(dsource, serviceName))

	method, err := dsource.FindSymbol(strings.Join([]string{serviceName, methodName}, "."))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%T\n", method)
	methd, ok := method.(*desc.MethodDescriptor)
	if !ok {
		panic("not a method")
	}

	reqBytes := []byte(`{"KbcName": "test"}`)
	req := methd.GetInputType().AsDescriptorProto()

	fmt.Println(req.GetName())
	fmt.Println(req.GetField())
	err = protojson.UnmarshalOptions{
		AllowPartial: true,
	}.Unmarshal(reqBytes, req)
	if err != nil {
		panic(err)
	}

	resp := methd.GetOutputType().AsProto()

	// meth, ok := method.(protoreflect.MethodDescriptor)
	// if !ok {
	// 	panic("not a method")
	// }
	// req := dynamicpb.NewMessage(meth.Input())
	// resp := dynamicpb.NewMessage(meth.Output())

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
