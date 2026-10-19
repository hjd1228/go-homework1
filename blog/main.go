// 个人博客系统后端：Gin + GORM + JWT。
//
// 运行方式：go run .（详细说明见 README.md）
package main

import (
	"log"

	"blog/internal/config"
	"blog/internal/database"
	"blog/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DBPath)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Printf("数据库就绪: %s", cfg.DBPath)

	engine := router.New(db, cfg)

	log.Printf("服务启动，监听 %s", cfg.Addr)
	if err := engine.Run(cfg.Addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
