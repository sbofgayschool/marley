package service

import (
	"github.com/gin-gonic/gin"
	"github.com/sbofgayschool/marley/server/service/common"
)

var server *gin.Engine

func init() {
	server = gin.Default()
	common.RegisterHandler(server)
}

func Run() {
	server.Run(":8081")
}
