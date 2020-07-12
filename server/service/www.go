package service

import (
	"github.com/gin-gonic/gin"

	"github.com/sbofgayschool/marley/server/service/common"

	_ "github.com/sbofgayschool/marley/server/service/chat"
	_ "github.com/sbofgayschool/marley/server/service/live"
)

var server *gin.Engine

func init() {
	server = gin.Default()
	common.RegisterHandler(server)

	server.Static("web", "web")
}

func Run() {
	_ = server.Run(":8081")
}
