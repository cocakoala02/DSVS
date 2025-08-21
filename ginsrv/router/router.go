package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ldsec/drynx/ginsrv/controller"
)

var ctl controller.Controller

// CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // ⚠️ 可替换成前端实际地址
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// 新建一个服务器端
func ServerSetup() *http.Server {
	r := gin.Default()

	// ✅ 添加 CORS 中间件
	r.Use(CORSMiddleware())

	// 注册路由
	RouterList(r)

	s := &http.Server{
		Addr:    "0.0.0.0:8088",
		Handler: r,
	}

	return s
}

// 地址列表及其处理函数
func RouterList(r *gin.Engine) {
	r.POST("/survey", ctl.Drynx.TripartiteSurvey)
}
