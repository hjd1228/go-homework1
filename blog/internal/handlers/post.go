package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"blog/internal/middleware"
	"blog/internal/models"
	"blog/internal/response"
)

// PostHandler 处理文章相关接口。
type PostHandler struct {
	db *gorm.DB
}

// NewPostHandler 构造 PostHandler。
func NewPostHandler(db *gorm.DB) *PostHandler {
	return &PostHandler{db: db}
}

type postRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required,min=1"`
}

// List 处理 GET /api/posts，支持 page、page_size 分页。
func (h *PostHandler) List(c *gin.Context) {
	page := parsePageQuery(c.Query("page"), 1)
	pageSize := parsePageQuery(c.Query("page_size"), 10)
	if pageSize > 100 {
		pageSize = 100
	}

	var total int64
	if err := h.db.Model(&models.Post{}).Count(&total).Error; err != nil {
		internalError(c, "统计文章数量失败", err)
		return
	}

	var posts []models.Post
	err := h.db.Preload("Author").
		Order("created_at DESC, id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&posts).Error
	if err != nil {
		internalError(c, "查询文章列表失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"items":     posts,
	})
}

// Get 处理 GET /api/posts/:id。
func (h *PostHandler) Get(c *gin.Context) {
	post, ok := h.findPost(c)
	if !ok {
		return
	}

	if err := h.db.Preload("Author").First(post, post.ID).Error; err != nil {
		internalError(c, "查询文章作者失败", err)
		return
	}

	c.JSON(http.StatusOK, post)
}

// Create 处理 POST /api/posts，需要登录。
func (h *PostHandler) Create(c *gin.Context) {
	var req postRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数不合法: "+err.Error())
		return
	}

	post := models.Post{
		Title:   req.Title,
		Content: req.Content,
		UserID:  middleware.CurrentUserID(c),
	}
	if err := h.db.Create(&post).Error; err != nil {
		internalError(c, "创建文章失败", err)
		return
	}

	// 回查一次，把作者信息一并返回
	if err := h.db.Preload("Author").First(&post, post.ID).Error; err != nil {
		internalError(c, "查询文章失败", err)
		return
	}

	c.JSON(http.StatusCreated, post)
}

// Update 处理 PUT /api/posts/:id，仅作者本人可修改。
func (h *PostHandler) Update(c *gin.Context) {
	post, ok := h.findPost(c)
	if !ok {
		return
	}

	if post.UserID != middleware.CurrentUserID(c) {
		response.Error(c, http.StatusForbidden, "只能修改自己的文章")
		return
	}

	var req postRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数不合法: "+err.Error())
		return
	}

	post.Title = req.Title
	post.Content = req.Content
	// Omit("Author")：post.Author 是零值，不排除的话 Save 会连带 upsert 一条假用户
	if err := h.db.Omit("Author").Save(post).Error; err != nil {
		internalError(c, "更新文章失败", err)
		return
	}

	// 补上作者信息再返回，避免响应里出现 id 为 0 的空 author
	if err := h.db.Preload("Author").First(post, post.ID).Error; err != nil {
		internalError(c, "查询文章失败", err)
		return
	}

	c.JSON(http.StatusOK, post)
}

// Delete 处理 DELETE /api/posts/:id，仅作者本人可删除。
func (h *PostHandler) Delete(c *gin.Context) {
	post, ok := h.findPost(c)
	if !ok {
		return
	}

	if post.UserID != middleware.CurrentUserID(c) {
		response.Error(c, http.StatusForbidden, "只能删除自己的文章")
		return
	}

	// 同一事务内先删评论再删文章，避免留下孤儿数据
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", post.ID).Delete(&models.Comment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Post{}, post.ID).Error
	})
	if err != nil {
		internalError(c, "删除文章失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文章已删除", "id": post.ID})
}

// findPost 按路径参数加载文章，失败时已写入响应。
func (h *PostHandler) findPost(c *gin.Context) (*models.Post, bool) {
	id, ok := parseIDParam(c)
	if !ok {
		return nil, false
	}

	var post models.Post
	if err := h.db.First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, "文章不存在")
			return nil, false
		}
		internalError(c, "查询文章失败", err)
		return nil, false
	}

	return &post, true
}

// parsePageQuery 解析分页参数，非法或缺失时回退到默认值。
func parsePageQuery(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}
	return value
}
