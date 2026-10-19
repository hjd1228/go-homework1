// Package middleware 提供 Gin 的通用中间件：请求日志与 JWT 认证。
package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 记录每个请求的方法、路径、状态码与耗时。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		log.Printf("[gin] %s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start).Round(time.Millisecond),
		)
	}
}
