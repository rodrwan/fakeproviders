package main

import (
	"flag"

	"github.com/rodrwan/fakeproviders/cmd/cards/server"
)

var (
	port  = flag.String("port", "8082", "Service port")
	token = flag.String("token", "fasdfadfa9fj987afsdf", "Token for authenticated endpointds")
)

func main() {
	flag.Parse()
	server.Run(*port)
}
