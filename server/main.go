package main

import (
	"github.com/sbofgayschool/marley/server/www"
)

func main() {
	www.Load().Run(":8081")
}
