package www

import (
	"github.com/gin-gonic/gin"
)

var server *gin.Engine

func init() {
	server = gin.Default()
}

func Run() {
	server.Run(":8081")
}
