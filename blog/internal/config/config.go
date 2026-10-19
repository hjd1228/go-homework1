// Package config 负责加载服务运行期的配置。
package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"time"
)

// Config 汇总服务运行所需的配置项。
type Config struct {
	// Addr 是 HTTP 服务监听地址。
	Addr string
	// DBPath 是 SQLite 数据库文件路径。
	DBPath string
	// JWTSecret 用于签发与校验 JWT。
	JWTSecret []byte
	// TokenTTL 是 JWT 的有效期。
	TokenTTL time.Duration
}

// Load 从环境变量读取配置，未设置的项使用开发默认值。
//
// 支持的变量：
//
//	JWT_SECRET  JWT 签名密钥（未设置时随机生成，仅适用于本地开发）
//	BLOG_ADDR   监听地址，默认 :8080
//	BLOG_DB     SQLite 文件路径，默认 blog.db
func Load() Config {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = randomSecret()
		log.Println("[warn] 未设置 JWT_SECRET，已随机生成临时密钥，服务重启后已签发的 token 将失效")
	}

	addr := os.Getenv("BLOG_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	dbPath := os.Getenv("BLOG_DB")
	if dbPath == "" {
		dbPath = "blog.db"
	}

	return Config{
		Addr:      addr,
		DBPath:    dbPath,
		JWTSecret: []byte(secret),
		TokenTTL:  24 * time.Hour,
	}
}

// randomSecret 生成 32 字节随机密钥的十六进制表示。
func randomSecret() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		log.Fatalf("生成 JWT 密钥失败: %v", err)
	}
	return hex.EncodeToString(buf)
}
