# ttRPCurl

`ttrpcurl` is a command-line tool that lets you interact with [ttRPC](https://github.com/containerd/ttrpc) servers. It's basically curl for ttRPC servers, and is heavily influenced by [gRPCurl](https://github.com/fullstorydev/grpcurl), which does the same thing for gRPC.

## Installation

Install via the Go command:

```sh
go install github.com/katexochen/ttrpcurl/cmd/ttrpcurl@latest
```

## Usage

The CLI comes with help flags, both global and on each command:

```sh
ttrpcurl --help
ttrpcurl <command> --help
```

## Syntax compatibility with gRPCurl

This tool tries to provide intuitive usability for people familiar with grpcurl. However, as ttrpcurl isn't using Go's native flag package there are some differences. Flags in ttrpcurl require two dashes (`--help`).

## TODOs

- [ ] Improve client-server testscript
- [ ] Support multiple proto files
- [ ] Write unit tests
- [ ] Write e2e tests
- [ ] Support streaming calls
- [ ] Support bidirectional streaming calls
- [ ] Support data from file with `@filename`
- [ ] Support protobuf text format
- [ ] Use timeout and other limits

## Limitations
