package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/pbaettig/20miner/internal/pkg/comments"
)

// "103117564"

func main() {
	cs := comments.GetComments("103117564")
	for _, c := range cs {
		spew.Dump(c)
	}

	fmt.Println(len(cs))
}
