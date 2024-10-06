package main

import (
	"flag"
	"fmt"

	scanintranet "github.com/tknie/scanIntranet"
)

const description = "Descrption"

func main() {
	create := false
	flag.BoolVar(&create, "C", false, "Create database")
	flag.Usage = func() {
		fmt.Print(description)
		fmt.Println("Default flags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	scanintranet.ScanIntranet(create)
}
