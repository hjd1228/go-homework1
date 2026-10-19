// Package response 统一接口的响应格式。
package response

import "github.com/gin-gonic/gin"

// Error 以统一结构返回错误，并终止后续处理。
func Error(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
