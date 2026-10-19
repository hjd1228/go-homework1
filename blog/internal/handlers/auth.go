package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"blog/internal/auth"
	"blog/internal/config"
	"blog/internal/models"
	"blog/internal/response"
)

// AuthHandler 处理用户注册与登录。
type AuthHandler struct {
	db  *gorm.DB
	cfg config.Config
}

// NewAuthHandler 构造 AuthHandler。
func NewAuthHandler(db *gorm.DB, cfg config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Email    string `json:"email" binding:"required,email"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register 处理 POST /api/register。
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数不合法: "+err.Error())
		return
	}

	// 用户名与邮箱都要求唯一，先查一次以便返回明确的冲突提示
	var existing models.User
	err := h.db.Where("username = ? OR email = ?", req.Username, req.Email).
		First(&existing).Error
	if err == nil {
		response.Error(c, http.StatusConflict, "用户名或邮箱已被注册")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		internalError(c, "查询用户失败", err)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		internalError(c, "密码加密失败", err)
		return
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashed),
		Email:    req.Email,
	}
	if err := h.db.Create(&user).Error; err != nil {
		internalError(c, "创建用户失败", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

// Login 处理 POST /api/login，成功后返回 JWT。
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数不合法: "+err.Error())
		return
	}

	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		// 不区分「用户不存在」与「密码错误」，避免账号枚举
		response.Error(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.Error(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	token, err := auth.GenerateToken(h.cfg.JWTSecret, user.ID, user.Username, h.cfg.TokenTTL)
	if err != nil {
		internalError(c, "生成 token 失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"expires_in": int(h.cfg.TokenTTL.Seconds()),
	})
}
