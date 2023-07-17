package main

import (
	"log"

	"github.com/katexochen/ttrpcurl/hack/testserver/grpctest"
)

func main() {
	if err := grpctest.Run(); err != nil {
		log.Fatal(err)
	}
}
