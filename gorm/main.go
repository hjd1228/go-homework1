// 进阶 GORM 练习：模型定义 / 关联查询 / 钩子函数
//
// 独立的 Go module，与仓库根的 homework01、advanced 互不影响。
// 运行方式：cd gorm && go run .
package main

import (
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// 题目 1：模型定义
// 假设博客系统有 User（用户）、Post（文章）、Comment（评论）三个实体，
// User 与 Post 是一对多（一个用户可发布多篇文章），
// Post 与 Comment 也是一对多（一篇文章可有多条评论）。
// ---------------------------------------------------------------------------

// User 用户。与 Post 是一对多关系。
type User struct {
	ID   uint
	Name string
	// PostCount 文章数量统计，由 Post 的 AfterCreate 钩子维护
	PostCount int `gorm:"default:0"`
	// Posts 一对多关联：一个用户拥有多篇文章
	Posts []Post `gorm:"foreignKey:UserID"`
}

// Post 文章。与 Comment 是一对多关系。
type Post struct {
	ID     uint
	UserID uint
	Title  string
	Body   string
	// CommentStatus 评论状态，由 Comment 的 AfterDelete 钩子在评论清空后写入「无评论」
	CommentStatus string `gorm:"default:有评论"`
	// Comments 一对多关联：一篇文章拥有多条评论
	Comments []Comment `gorm:"foreignKey:PostID"`
	// CommentCount 只用于承载聚合查询结果：
	// ->        只读，写库时忽略
	// -:migration 不参与建表，所以表里没有这一列
	CommentCount int64 `gorm:"->;-:migration"`
}

// Comment 评论。
type Comment struct {
	ID      uint
	PostID  uint
	Content string
}

// AutoMigrate 使用 GORM 创建 User、Post、Comment 对应的数据库表。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Post{}, &Comment{})
}

// ---------------------------------------------------------------------------
// 题目 2：关联查询
// 基于上面的博客系统模型。
// ---------------------------------------------------------------------------

// GetUserPostsWithComments 查询指定用户发布的所有文章及其对应的评论信息。
func GetUserPostsWithComments(db *gorm.DB, userID uint) ([]Post, error) {
	var posts []Post
	err := db.Preload("Comments").
		Where("user_id = ?", userID).
		Order("id ASC").
		Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// GetMostCommentedPost 查询评论数量最多的文章。
func GetMostCommentedPost(db *gorm.DB) (*Post, error) {
	var post Post
	// LEFT JOIN 保证没有评论的文章也会被统计（评论数为 0）
	err := db.Model(&Post{}).
		Select("posts.*, COUNT(comments.id) AS comment_count").
		Joins("LEFT JOIN comments ON comments.post_id = posts.id").
		Group("posts.id").
		Order("comment_count DESC, posts.id ASC").
		Limit(1).
		Scan(&post).Error
	if err != nil {
		return nil, err
	}
	if post.ID == 0 {
		return nil, nil
	}
	return &post, nil
}

// ---------------------------------------------------------------------------
// 题目 3：钩子函数
// 继续使用博客系统的模型。
// ---------------------------------------------------------------------------

// AfterCreate 在文章创建时自动更新所属用户的文章数量统计字段。
func (p *Post) AfterCreate(tx *gorm.DB) error {
	return tx.Model(&User{}).
		Where("id = ?", p.UserID).
		UpdateColumn("post_count", gorm.Expr("post_count + 1")).
		Error
}

// AfterDelete 在评论删除时检查文章的评论数量，
// 如果评论数量为 0，则把文章的评论状态更新为「无评论」。
func (c *Comment) AfterDelete(tx *gorm.DB) error {
	var remaining int64
	if err := tx.Model(&Comment{}).
		Where("post_id = ?", c.PostID).
		Count(&remaining).Error; err != nil {
		return err
	}

	if remaining > 0 {
		return nil
	}

	return tx.Model(&Post{}).
		Where("id = ?", c.PostID).
		Update("comment_status", "无评论").
		Error
}

// ---------------------------------------------------------------------------
// 入口
// ---------------------------------------------------------------------------

func main() {
	// 每次运行都从干净的数据库开始，避免历史数据干扰演示结果
	_ = os.Remove("blog.db")

	db, err := gorm.Open(sqlite.Open("blog.db"), &gorm.Config{})
	if err != nil {
		fmt.Println("打开数据库失败:", err)
		return
	}

	// ---- 题目 1：建表 ----
	if err := AutoMigrate(db); err != nil {
		fmt.Println("[1] 建表失败:", err)
		return
	}
	fmt.Println("[1] 建表完成：users / posts / comments")

	// ---- 造测试数据 ----
	user := User{Name: "张三"}
	if err := db.Create(&user).Error; err != nil {
		fmt.Println("创建用户失败:", err)
		return
	}

	posts := []Post{
		{UserID: user.ID, Title: "GORM 入门", Body: "模型定义与关联"},
		{UserID: user.ID, Title: "Go 并发", Body: "goroutine 与 channel"},
	}
	if err := db.Create(&posts).Error; err != nil {
		fmt.Println("创建文章失败:", err)
		return
	}

	comments := []Comment{
		{PostID: posts[0].ID, Content: "写得不错"},
		{PostID: posts[0].ID, Content: "学到了"},
		{PostID: posts[0].ID, Content: "期待下一篇"},
		{PostID: posts[1].ID, Content: "受教了"},
	}
	if err := db.Create(&comments).Error; err != nil {
		fmt.Println("创建评论失败:", err)
		return
	}

	// ---- 题目 3（钩子）：文章创建后文章数量已被自动维护 ----
	var refreshed User
	if err := db.First(&refreshed, user.ID).Error; err != nil {
		fmt.Println("查询用户失败:", err)
		return
	}
	fmt.Printf("[3] 创建 %d 篇文章后，%s 的文章数量统计: %d\n",
		len(posts), refreshed.Name, refreshed.PostCount)

	// ---- 题目 2：关联查询 ----
	userPosts, err := GetUserPostsWithComments(db, user.ID)
	if err != nil {
		fmt.Println("[2] 查询用户文章失败:", err)
		return
	}
	fmt.Printf("[2] %s 共 %d 篇文章:\n", refreshed.Name, len(userPosts))
	for _, p := range userPosts {
		fmt.Printf("    《%s》 %d 条评论\n", p.Title, len(p.Comments))
		for _, c := range p.Comments {
			fmt.Println("      -", c.Content)
		}
	}

	most, err := GetMostCommentedPost(db)
	if err != nil {
		fmt.Println("[2] 查询热门文章失败:", err)
		return
	}
	if most == nil {
		fmt.Println("[2] 暂无文章")
	} else {
		fmt.Printf("[2] 评论最多的文章：《%s》，共 %d 条评论\n", most.Title, most.CommentCount)
	}

	// ---- 题目 3（钩子）：删空某篇文章的评论，观察评论状态 ----
	target := posts[1]
	for _, c := range comments {
		if c.PostID != target.ID {
			continue
		}
		// 必须先查出来再删：钩子要靠 PostID 定位所属文章，
		// 而 db.Delete(&Comment{}, id) 只会带上主键，不含 PostID
		var loaded Comment
		if err := db.First(&loaded, c.ID).Error; err != nil {
			fmt.Println("查询评论失败:", err)
			return
		}
		if err := db.Delete(&loaded).Error; err != nil {
			fmt.Println("删除评论失败:", err)
			return
		}
	}

	var updated Post
	if err := db.First(&updated, target.ID).Error; err != nil {
		fmt.Println("查询文章失败:", err)
		return
	}
	fmt.Printf("[3] 删空《%s》的评论后，评论状态: %s\n", updated.Title, updated.CommentStatus)
}
