package main

import (
	"fmt"
	"strconv"
)

type KV struct {
	K string
	V string
}

type O struct {
	Name string
	Vals []KV
}

func main() {
	o := O{
		"test",
		[]KV{{"K1", ""}, {"K2", ""}, {"K3", ""}},
	}
	fmt.Printf("%+v\n", o.Vals)
	for i := range o.Vals {
		o.Vals[i].V = strconv.Itoa(i)
	}

	fmt.Printf("%+v\n", o.Vals)

}
