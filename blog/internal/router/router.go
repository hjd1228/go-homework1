// Package router 负责组装 Gin 引擎与路由表。
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"blog/internal/config"
	"blog/internal/handlers"
	"blog/internal/middleware"
	"blog/internal/response"
)

// New 构建带有全部路由与中间件的 Gin 引擎。
func New(db *gorm.DB, cfg config.Config) *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.Logger(), gin.Recovery())

	engine.NoRoute(func(c *gin.Context) {
		response.Error(c, http.StatusNotFound, "接口不存在")
	})

	authHandler := handlers.NewAuthHandler(db, cfg)
	postHandler := handlers.NewPostHandler(db)
	commentHandler := handlers.NewCommentHandler(db)

	api := engine.Group("/api")
	{
		// 公开接口
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		api.GET("/posts", postHandler.List)
		api.GET("/posts/:id", postHandler.Get)
		api.GET("/posts/:id/comments", commentHandler.List)

		// 需要登录的接口
		authed := api.Group("")
		authed.Use(middleware.Auth(cfg.JWTSecret))
		{
			authed.POST("/posts", postHandler.Create)
			authed.PUT("/posts/:id", postHandler.Update)
			authed.DELETE("/posts/:id", postHandler.Delete)
			authed.POST("/posts/:id/comments", commentHandler.Create)
		}
	}

	return engine
}
