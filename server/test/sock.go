package main

import (
    "log"

    "github.com/gin-gonic/gin"

    "github.com/sbofgayschool/marley/server/infra/sock"
)

func init() {
    sock.RegisterHandler("test", func(message *sock.Message, messages chan *sock.Message) []*sock.Message {
        return []*sock.Message{&sock.Message{Client: nil, Content: map[string]string{"Content": "pong"}}}
    })
}

func main() {
    go sock.Run()
    g := gin.New()
    g.GET("/:id", func(c *gin.Context) {
        id := c.Param("id")
        if err := sock.NewClient(c, id, nil); err != nil {
            log.Println(err)
        }
    })
    g.Run(":2808")
}
