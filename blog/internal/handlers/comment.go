package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"blog/internal/middleware"
	"blog/internal/models"
	"blog/internal/response"
)

// CommentHandler 处理评论相关接口。
type CommentHandler struct {
	db *gorm.DB
}

// NewCommentHandler 构造 CommentHandler。
func NewCommentHandler(db *gorm.DB) *CommentHandler {
	return &CommentHandler{db: db}
}

type commentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

// Create 处理 POST /api/posts/:id/comments，需要登录。
func (h *CommentHandler) Create(c *gin.Context) {
	post, ok := h.findPost(c)
	if !ok {
		return
	}

	var req commentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数不合法: "+err.Error())
		return
	}

	comment := models.Comment{
		Content: req.Content,
		PostID:  post.ID,
		UserID:  middleware.CurrentUserID(c),
	}
	if err := h.db.Create(&comment).Error; err != nil {
		internalError(c, "创建评论失败", err)
		return
	}

	if err := h.db.Preload("Author").First(&comment, comment.ID).Error; err != nil {
		internalError(c, "查询评论失败", err)
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// List 处理 GET /api/posts/:id/comments。
func (h *CommentHandler) List(c *gin.Context) {
	post, ok := h.findPost(c)
	if !ok {
		return
	}

	var comments []models.Comment
	err := h.db.Preload("Author").
		Where("post_id = ?", post.ID).
		Order("created_at ASC, id ASC").
		Find(&comments).Error
	if err != nil {
		internalError(c, "查询评论列表失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": len(comments), "items": comments})
}

// findPost 校验路径参数对应的文章存在，失败时已写入响应。
func (h *CommentHandler) findPost(c *gin.Context) (*models.Post, bool) {
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
