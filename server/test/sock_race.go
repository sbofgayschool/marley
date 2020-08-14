package main

import (
    "fmt"
    "log"
    "sync"

    "github.com/gin-gonic/gin"

    "github.com/sbofgayschool/marley/server/infra/sock"
)

var data = make(map[string]*int)
var lock = sync.Mutex{}

func init() {
    data["1"] = new(int)
    data["2"] = new(int)
    sock.RegisterHandler("test", func(message *sock.Message, messages chan *sock.Message) []*sock.Message {
        lock.Lock()
        dt := data[message.Client.Gid]
        data[message.Client.Gid] = dt
        lock.Unlock()
        *dt++
        return []*sock.Message{&sock.Message{Client: nil, Content: message.Content}}
    })
}

func main() {
    defer func() {
        fmt.Println(*data["1"])
        fmt.Println(*data["2"])
    }()
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
