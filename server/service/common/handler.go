package common

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/sbofgayschool/marley/server/infra/sock"
	"github.com/sbofgayschool/marley/server/service/user"
	"github.com/sbofgayschool/marley/server/utils"
)

const (
	FileUploadDir = "web/res/file/"
)

func RegisterHandler(engine *gin.Engine) {
	engine.GET("api/test", func(c *gin.Context) {
		c.String(200, "Marley server running.")
	})
	engine.GET("api/sock/:id", UpgradeHandler)
	engine.POST("api/file", UploadFileHandler)
}

func UpgradeHandler(c *gin.Context) {
	id := c.Param("id")
	// TODO: Fetch uid from path parameters and authorize the user.
	if err := sock.NewClient(c, id, &user.SockUser{Uid: 0, Username: "Anonymous User", Teacher: true}); err != nil {
	}
}

func UploadFileHandler(c *gin.Context) {
	file, _ := c.FormFile("file")
	names := strings.Split(file.Filename, ".")
	fileType := "file"
	switch names[len(names)-1] {
	case "jpg": fallthrough
	case "png": fallthrough
	case "gif": fileType = "image"
	case "mp3": fallthrough
	case "wav": fallthrough
	case "ogg": fileType = "audio"
	}
	filename := utils.RandomString() + "." + names[len(names)-1]
	_ = c.SaveUploadedFile(file, FileUploadDir+filename)
	c.JSON(200, gin.H{"File": filename, "Type": fileType})
}


