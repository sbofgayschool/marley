package service

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/sbofgayschool/marley/server/service/course"
	"github.com/sbofgayschool/marley/server/service/user"
	"github.com/sbofgayschool/marley/server/service/vod"

	"github.com/sbofgayschool/marley/server/service/common"

	_ "github.com/sbofgayschool/marley/server/service/chat"
	_ "github.com/sbofgayschool/marley/server/service/live"
	_ "github.com/sbofgayschool/marley/server/service/user"
)

var server *gin.Engine

func init() {
	server = gin.Default()
	server.Use(sessions.Sessions("marley", cookie.NewStore([]byte("secret"))))

	common.RegisterHandler(server)
	user.RegisterHandler(server)
	course.RegisterHandler(server)
	vod.RegisterHandler(server)

	server.Static("web", "web")
}

func Run() {
	_ = server.Run(":8081")
}
