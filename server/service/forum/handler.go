package forum

import (
	"github.com/gin-gonic/gin"
	"github.com/sbofgayschool/marley/server/service/course"
	"github.com/sbofgayschool/marley/server/service/user"
	"strconv"
)

func AddForumHandler(c *gin.Context) {
	u := c.MustGet("user").(*user.User)
	title := c.DefaultPostForm("title", "")
	content := c.DefaultPostForm("content", "")
	if courseId, err := strconv.Atoi(c.DefaultPostForm("course", "")); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else if res, err := course.GetCourse(courseId, u.Id); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else if res.Relation == 0 && res.Owner != u.Id {
		c.JSON(403, gin.H{"Message": "operation not allowed"})
		return
	} else if res, err := AddForum(u.Id, courseId, title, content); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else {
		c.JSON(200, gin.H{"Id": res})
	}
}

func GetForumHandler(c *gin.Context) {
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else if res, err := GetForum(id); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else {
		c.JSON(200, res)
	}
}

func SearchForumHandler(c *gin.Context) {
	if courseId, err := strconv.Atoi(c.DefaultQuery("course", "")); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else if res, err := SearchForum(courseId); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else {
		c.JSON(200, res)
	}
}

func DeleteForumHandler(c *gin.Context) {
	u := c.MustGet("user").(*user.User)
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else if res, err := GetForum(id); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else {
		if res.User != u.Id {
			if res, err := course.GetCourse(res.Course, u.Id); err != nil {
				c.JSON(500, gin.H{"Message": err.Error()})
				return
			} else if res.Relation < 2 && res.Owner != u.Id {
				c.JSON(403, gin.H{"Message": "operation not allowed"})
				return
			}
		}
		if err := DeleteForum(id); err != nil {
			c.JSON(500, gin.H{"Message": err.Error()})
			return
		}
	}
	c.JSON(200, gin.H{})
}

func AddReplyHandler(c *gin.Context) {
	u := c.MustGet("user").(*user.User)
	content := c.DefaultPostForm("content", "")
	if forum, err := strconv.Atoi(c.Param("id")); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else if f, err := GetForum(forum); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else if res, err := course.GetCourse(f.Course, u.Id); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else if res.Relation == 0 && res.Owner != u.Id {
		c.JSON(403, gin.H{"Message": "operation not allowed"})
		return
	} else if err := AddReply(u.Id, forum, content); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	}
	c.JSON(200, gin.H{})
}

func SearchReplyHandler(c *gin.Context) {
	if forum, err := strconv.Atoi(c.Param("id")); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else if res, err := SearchReply(forum); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else {
		c.JSON(200, res)
	}
}

func DeleteReplyHandler(c *gin.Context) {
	u := c.MustGet("user").(*user.User)
	if id, err := strconv.Atoi(c.Param("reply")); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else if res, err := GetReply(id); err != nil {
		c.JSON(500, gin.H{"Message": err.Error()})
		return
	} else {
		if res.User != u.Id {
			if f, err := GetForum(res.Forum); err != nil {
				c.JSON(500, gin.H{"Message": err.Error()})
				return
			} else if res, err := course.GetCourse(f.Course, u.Id); err != nil {
				c.JSON(500, gin.H{"Message": err.Error()})
				return
			} else if res.Relation < 2 && res.Owner != u.Id {
				c.JSON(403, gin.H{"Message": "operation not allowed"})
				return
			}
		}
		if err := DeleteReply(id); err != nil {
			c.JSON(500, gin.H{"Message": err.Error()})
			return
		}
	}
	c.JSON(200, gin.H{})
}

func RegisterHandler(engine *gin.Engine) {
	f := engine.Group("api/forum/")
	f.Use(user.GetCurrentUserMiddleware())
	f.PUT("forum/add", AddForumHandler)
	f.GET("forum/get/:id", GetForumHandler)
	f.GET("forum/search", SearchForumHandler)
	f.DELETE("forum/delete/:id", DeleteForumHandler)
	f.PUT("reply/add/:id", AddReplyHandler)
	f.GET("reply/search/:id", SearchReplyHandler)
	f.DELETE("reply/delete/:id/:reply", DeleteReplyHandler)
}
