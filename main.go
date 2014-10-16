package main

import "flag"

var serverAddress = flag.String(
	"address",
	"",
	"Specifies the address to bind to",
)

func main() {
	flag.Parse()
}
