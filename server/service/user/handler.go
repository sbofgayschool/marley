package user

import (
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/sbofgayschool/marley/server/utils"
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
		c.JSON(409, gin.H{"Message": "user with same username exists"})
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
		c.JSON(401, gin.H{"Message": "incorrect username or password"})
		return
	}
	s := sessions.Default(c)
	s.Set("userId", user.Id)
	s.Set("userUsername", user.Username)
	s.Set("userTeacher", user.Teacher)
	s.Set("userNote", user.Note)
	s.Save()
	c.JSON(200, gin.H{})
}

func GetCurrentUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.Default(c)
		u := s.Get("userId")
		if u == nil {
			c.JSON(401, gin.H{"Message": "login required"})
			c.Abort()
			return
		}
		c.Set("user", &User{
			Id:       u.(int),
			Username: s.Get("userUsername").(string),
			Teacher:  s.Get("userTeacher").(int),
			Note:     s.Get("userNote").(string),
		})
	}
}

func LogoutHandler(c *gin.Context) {
	sessions.Default(c).Delete("userId")
	sessions.Default(c).Save()
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
	sessions.Default(c).Set("userNote", note)
	sessions.Default(c).Save()
	c.JSON(200, user)
}

func SetUserPasswordHandler(c *gin.Context) {
	originalPassword := utils.GetHash(c.DefaultPostForm("originalPassword", ""))
	password := utils.GetHash(c.DefaultPostForm("password", ""))
	user := c.MustGet("user").(*User)
	if _, p, err := GetUser(user.Id, ""); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else if p != originalPassword {
		c.JSON(403, gin.H{"Message": "incorrect password"})
		return
	}
	if err := SetPassword(user.Id, password); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
	}
	c.JSON(200, gin.H{})
}

func SearchUserHandler(c *gin.Context) {
	res, err := SearchUser()
	if err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
	}
	c.JSON(200, res)
}

func RegisterHandler(engine *gin.Engine) {
	engine.PUT("api/register", RegHandler)
	engine.POST("api/login", LoginHandler)
	engine.GET("api/logout", LogoutHandler)

	u := engine.Group("api/user/")
	u.Use(GetCurrentUserMiddleware())
	u.GET("get", GetUserHandler)
	u.POST("note", SetUserNoteHandler)
	u.POST("password", SetUserPasswordHandler)
	u.GET("search", SearchUserHandler)
}
