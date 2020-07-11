package common

import (
    "github.com/gin-gonic/gin"
    "github.com/sbofgayschool/marley/server/infra/sock"
    "github.com/sbofgayschool/marley/server/service/user"
)

func Upgrade(c *gin.Context) {
    id := c.Param("id")
    // TODO: Fetch uid from middleware and authorize the user.
    if err := sock.NewClient(c, id, &user.SockUser{Uid: 0, Username: "Anonymous User", Teacher: true}); err != nil {
    }
}

func RegisterHandler(engine *gin.Engine) {
    engine.GET("api/sock/:id", Upgrade)
}
