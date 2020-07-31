package course

import (
    "log"
    "strconv"

    "github.com/gin-gonic/gin"

    "github.com/sbofgayschool/marley/server/infra/sock"
    // "github.com/sbofgayschool/marley/server/service/common"
    "github.com/sbofgayschool/marley/server/service/user"
)

func UpgradeHandler(c *gin.Context) {
    u := c.MustGet("user").(*user.User)
    log.Println(u)
    id := c.Param("id")
    /*
    cid, _ := common.GetIdVodId(id)
    cidInt, _ := strconv.Atoi(cid)
    course, _ := GetCourse(cidInt, u.Id)
    teacher := course.Owner == u.Id || course.Relation == 2
    */
    teacher := true
    if err := sock.NewClient(c, id, &user.SockUser{Uid: u.Id, Username: u.Username, Teacher: teacher}); err != nil {
        log.Println(err)
    }
}

func AddCourseHandler(c *gin.Context) {
    u := c.MustGet("user").(*user.User)
    name := c.DefaultPostForm("name", "")
    tag := c.DefaultPostForm("tag", "")
    note := c.DefaultPostForm("note", "")
    if u.Teacher == 0 {
        c.JSON(403, gin.H{"Message": "operation not allowed"})
        return
    } else if res, err := AddCourse(name, u.Id, tag, note); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else {
        c.JSON(200, gin.H{"Id": res})
    }
}

func GetCourseHandler(c *gin.Context) {
    u := c.MustGet("user").(*user.User)
    if id, err := strconv.Atoi(c.Param("id")); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else if res, err := GetCourse(id, u.Id); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else {
        c.JSON(200, res)
    }
}

func SearchCourseHandler(c *gin.Context) {
    u := c.MustGet("user").(*user.User)
    if res, err := SearchCourse(u.Id); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else {
        c.JSON(200, res)
    }
}

func SetCourseHandler(c *gin.Context) {
    u := c.MustGet("user").(*user.User)
    tag := c.DefaultPostForm("tag", "")
    note := c.DefaultPostForm("note", "")
    if id, err := strconv.Atoi(c.Param("id")); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else if res, err := GetCourse(id, u.Id); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else if res.Owner != u.Id {
        c.JSON(403, gin.H{"Message": "operation not allowed"})
        return
    } else if err := SetCourse(id, tag, note); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    }
    c.JSON(200, gin.H{})
}

func SearchRelationHandler(c *gin.Context) {
    if id, err := strconv.Atoi(c.Param("id")); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else if res, err := SearchRelation(id); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else {
        c.JSON(200, res)
    }
}

func SetRelationHandler(c *gin.Context) {
    u := c.MustGet("user").(*user.User)
    if id, err := strconv.Atoi(c.Param("id")); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else if target, err := strconv.Atoi(c.DefaultPostForm("user", "0")); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else if course, err := GetCourse(id, u.Id); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else {
        if course.Owner != u.Id {
            if course.Relation == 0 {
                err = SetRelation(id, u.Id, 1)
            } else {
                err = SetRelation(id, u.Id, 0)
            }
        } else if target == u.Id {
            c.JSON(409, gin.H{"Message": "target user not allowed"})
            return
        } else if tu, _,  e := user.GetUser(target, ""); e != nil {
            c.JSON(500, gin.H{"Message": e.Error()})
            return
        } else if tu == nil {
            c.JSON(409, gin.H{"Message": "target user not found"})
            return
        } else if course, err = GetCourse(id, target); err == nil {
            if course.Relation == 0 {
                err = SetRelation(id, target, 1 + tu.Teacher)
            } else {
                err = SetRelation(id, target, 0)
            }
        }
        if err != nil {
            c.JSON(500, gin.H{"Message": err.Error()})
        } else {
            c.JSON(200, gin.H{})
        }
    }
}

func AddCommentHandler(c *gin.Context) {
    u := c.MustGet("user").(*user.User)
    comment := c.DefaultPostForm("comment", "")
    if id, err := strconv.Atoi(c.Param("id")); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else if rate, err := strconv.Atoi(c.DefaultPostForm("rate", "0")); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else if err := AddComment(id, u.Id, rate, comment); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    }
    c.JSON(200, gin.H{})
}

func SearchCommentHandler(c *gin.Context) {
    if id, err := strconv.Atoi(c.Param("id")); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else if res, err := SearchComment(id); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else {
        c.JSON(200, res)
    }
}

func DeleteCommentHandler(c *gin.Context) {
    u := c.MustGet("user").(*user.User)
    if id, err := strconv.Atoi(c.Param("id")); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    } else if err := DeleteComment(id, u.Id); err != nil {
        c.JSON(500, gin.H{"Message": err.Error()})
        return
    }
    c.JSON(200, gin.H{})
}

func RegisterHandler(engine *gin.Engine) {
    s := engine.Group("api/sock/")
    s.Use(user.GetCurrentUserMiddleware())
    s.GET(":id", UpgradeHandler)

    c := engine.Group("api/course/")
    c.Use(user.GetCurrentUserMiddleware())
    c.PUT("course/add", AddCourseHandler)
    c.GET("course/get/:id", GetCourseHandler)
    c.GET("course/search", SearchCourseHandler)
    c.POST("course/set/:id", SetCourseHandler)
    c.GET("relation/search/:id", SearchRelationHandler)
    c.POST("relation/set/:id", SetRelationHandler)
    c.PUT("comment/add/:id", AddCommentHandler)
    c.GET("comment/search/:id", SearchCommentHandler)
    c.DELETE("comment/delete/:id", DeleteCommentHandler)
}
