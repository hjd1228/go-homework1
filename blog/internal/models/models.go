// Package models 定义博客系统的数据库模型。
package models

import "time"

// User 用户表。
//
// Password 字段带 json:"-"，保证密码哈希永远不会出现在接口响应里。
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password  string    `gorm:"size:100;not null" json:"-"`
	Email     string    `gorm:"uniqueIndex;size:128;not null" json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Post 文章表，通过 UserID 关联 User。
type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Author    User      `gorm:"foreignKey:UserID" json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Comment 评论表，同时关联 User 与 Post。
type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Author    User      `gorm:"foreignKey:UserID" json:"author"`
	PostID    uint      `gorm:"index;not null" json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
