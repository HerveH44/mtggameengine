package routers

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	r.Use(static.Serve("/", static.LocalFile("/home/noname/projects/dr4ft/built", false)))

	return r
}
