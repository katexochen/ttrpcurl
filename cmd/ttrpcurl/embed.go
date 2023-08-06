//go:build embed

package main

import "embed"

//go:embed protoinclude/*.proto
var protoIncludeFS embed.FS

const (
	protoIncludePath  = "protoinclude"
	protoFlagRequired = false
)
