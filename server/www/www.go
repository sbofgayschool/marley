package www

import (
	"github.com/gin-gonic/gin"
)

func Load() *gin.Engine {
	return gin.Default()
}
