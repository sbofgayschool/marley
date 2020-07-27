package user

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sbofgayschool/marley/server/infra/sock"
	"github.com/sbofgayschool/marley/server/utils"
	"log"
	"strconv"
)

func RegHandler(c *gin.Context) {
	username := c.DefaultPostForm("username", "")
	password := utils.GetHash(c.DefaultPostForm("password", ""))
	teacher, _ := strconv.Atoi(c.DefaultPostForm("teacher", "0"))
	note := c.DefaultPostForm("note", "")

	if user, _, err := GetUser(-1, username); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else if user != nil {
		c.JSON(500, gin.H{"Message": "user with same username exists"})
		return
	}
	if err := AddUser(username, password, teacher, note); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
	}
	c.JSON(200, gin.H{})
}

func LoginHandler(c *gin.Context) {
	if _, ok := c.Get("user"); ok {
		c.JSON(200, gin.H{})
		return
	}
	username := c.DefaultPostForm("username", "")
	password := utils.GetHash(c.DefaultPostForm("password", ""))

	user, p, err := GetUser(-1, username)
	if err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else if user == nil || password != p {
		c.JSON(500, gin.H{"Message": "incorrect username or password"})
		return
	}
	sessions.Default(c).Set("user", user)
	sessions.Default(c).Save()
	c.JSON(200, gin.H{})
}

func GetCurrentUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := sessions.Default(c).Get("user")
		if u == nil {
			c.JSON(401, gin.H{"message": "login required"})
			c.Abort()
			return
		}
		c.Set("user", u)
	}
}

func LogoutHandler(c *gin.Context) {
	sessions.Default(c).Delete("user")
	c.JSON(200, gin.H{})
}

func GetUserHandler(c *gin.Context) {
	id, _ := strconv.Atoi(c.DefaultQuery("id", "-1"))
	if id == -1 {
		c.JSON(200, c.MustGet("user"))
		return
	}
	user, _, err := GetUser(id, "")
	if err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	}
	c.JSON(200, user)
}

func SetUserNoteHandler(c *gin.Context) {
	note := c.DefaultPostForm("note", "")
	user := c.MustGet("user").(*User)
	if err := SetNote(user.Id, note); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	}
	user.Note = note
	c.JSON(200, user)
}

func SetUserPasswordHandler(c *gin.Context) {
	password := utils.GetHash(c.DefaultPostForm("password", ""))
	user := c.MustGet("user").(*User)
	if _, p, err := GetUser(user.Id, ""); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else if p != password {
		c.JSON(403, gin.H{"Message": "incorrect password"})
		return
	}
	if err := SetPassword(user.Id, password); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
	}
	c.JSON(200, gin.H{})
}

func SearchUserHandler(c *gin.Context) {
	username := c.DefaultPostForm("username", "")
	teacher, _ := strconv.Atoi(c.DefaultPostForm("teacher", "-1"))
	res, err := SearchUser(username, teacher)
	if err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
	}
	c.JSON(200, res)
}

func UpgradeHandler(c *gin.Context) {
	log.Println(sessions.Default(c).Get("user"))
	id := c.Param("id")
	// TODO: Fetch uid from path parameters and authorize the user.
	if err := sock.NewClient(c, id, &SockUser{Uid: 0, Username: "Anonymous User", Teacher: true}); err != nil {
	}
}

func RegisterHandler(engine *gin.Engine) {
	engine.GET("api/sock/:id", UpgradeHandler)
	engine.POST("api/register", RegHandler)
	engine.POST("api/login", LoginHandler)
	engine.GET("api/logout", LogoutHandler)

	r := engine.Group("api/user/")
	r.Use(GetCurrentUserMiddleware())
	r.GET("get", GetUserHandler)
	r.POST("note", SetUserNoteHandler)
	r.POST("password", SetUserPasswordHandler)
	r.GET("search", SearchUserHandler)
}
