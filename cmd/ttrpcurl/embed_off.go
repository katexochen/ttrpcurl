//go:build !embed

package main

import "embed"

var protoIncludeFS embed.FS

const (
	protoIncludePath  = "."
	protoFlagRequired = true
)
