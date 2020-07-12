package main

import (
	"github.com/sbofgayschool/marley/server/infra/sock"
	"github.com/sbofgayschool/marley/server/service"
)

func main() {
	go sock.Run()
	service.Run()
}
