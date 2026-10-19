// Package handlers 提供各接口的处理函数。
package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"blog/internal/response"
)

// parseIDParam 解析路径参数 :id，失败时已写入响应。
func parseIDParam(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		response.Error(c, http.StatusBadRequest, "id 不合法")
		return 0, false
	}
	return uint(value), true
}

// internalError 记录错误日志并返回 500，避免把底层错误细节暴露给调用方。
func internalError(c *gin.Context, message string, err error) {
	log.Printf("[error] %s: %v", message, err)
	response.Error(c, http.StatusInternalServerError, message)
}
