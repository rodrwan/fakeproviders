package main

import (
	"flag"

	"github.com/rodrwan/fakeproviders/cmd/users/server"
)

var (
	port  = flag.String("port", "8080", "Service port")
	token = flag.String("token", "fasdfadfa9fj987afsdf", "Token for authenticated endpointds")
)

func main() {
	flag.Parse()

	server.Run(*port)
}
