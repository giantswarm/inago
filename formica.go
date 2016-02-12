package main

import (
	"fmt"

	"github.com/giantswarm/formica/fleet"
)

func main() {
	newFleet, err := fleet.NewFleet(fleet.DefaultConfig())
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", newFleet)
}
