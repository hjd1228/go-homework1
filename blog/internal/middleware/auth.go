package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"blog/internal/auth"
	"blog/internal/response"
)

// contextUserID 是认证通过后写入 gin.Context 的用户 ID 键名。
const contextUserID = "userID"

// Auth 校验 Authorization: Bearer <token>，通过后把用户 ID 写入请求上下文。
func Auth(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Error(c, http.StatusUnauthorized, "缺少 Authorization 请求头")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			response.Error(c, http.StatusUnauthorized, "Authorization 请求头格式应为 Bearer <token>")
			return
		}

		claims, err := auth.ParseToken(secret, parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "token 无效或已过期")
			return
		}

		c.Set(contextUserID, claims.UserID)
		c.Next()
	}
}

// CurrentUserID 取出认证中间件写入的用户 ID；未认证时返回 0。
func CurrentUserID(c *gin.Context) uint {
	value, exists := c.Get(contextUserID)
	if !exists {
		return 0
	}
	userID, _ := value.(uint)
	return userID
}
