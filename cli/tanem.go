package main

import (
	"os"
	//"fmt"
	"github.com/ii64/tanem/cmd"
)

func main() {
	if err := cmd.NewTanemCmd(os.Args); err != nil {
		//fmt.Printf("[ERR]: %s", err)
		panic(err)
	}
}