// Package database 负责数据库连接与表结构迁移。
package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"blog/internal/models"
)

// Connect 打开 SQLite 数据库连接。
//
// GORM 的默认日志会把「记录不存在」这类正常业务分支也打印出来，
// 这里改成只记录慢查询与错误，保持日志可读。
func Connect(dsn string) (*gorm.DB, error) {
	logger := gormlogger.New(
		log.New(os.Stdout, "[gorm] ", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	return gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger})
}

// Migrate 按外键依赖顺序迁移 User、Post、Comment 三张表。
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{})
}
