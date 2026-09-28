package service

import (
	"errors"
	"net/http"
	"paopao/internal/auth"
	"paopao/internal/db"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserRegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserHandler struct {
	db        *gorm.DB
	JWTSecret string
}

func NewUserHandler(db *gorm.DB, jwt string) *UserHandler {
	return &UserHandler{db: db, JWTSecret: jwt}
}

func (u *UserHandler) Register(c *gin.Context) {
	var UserRegister UserRegisterRequest
	err := c.ShouldBindJSON(&UserRegister)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析失败"})
		return
	}
	var existingUser db.User
	if err := u.db.Where("email = ?", UserRegister.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "用户已存在"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	data, err := bcrypt.GenerateFromPassword([]byte(UserRegister.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码生成失败"})
		return
	}
	user := db.User{
		Name:     UserRegister.Name,
		Email:    UserRegister.Email,
		Password: string(data),
		Role:     "customer",
	}
	if err := u.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "注册成功",
		"id":     user.ID,
		"name":   user.Name,
		"email":  user.Email,
	})
}

func (u *UserHandler) Login(c *gin.Context) {
	var UserRegister UserLoginRequest
	err := c.ShouldBindJSON(&UserRegister)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析失败"})
		return
	}
	var existingUser db.User
	if err := u.db.Where("email = ?", UserRegister.Email).First(&existingUser).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户不存在，请先注册"})
		return
	}
	// 查到记录，邮箱已存在
	if err1 := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(UserRegister.Password)); err1 != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码错误"})
		return
	}
	//密码正确，签发token
	tokenString, err := auth.Sign(&existingUser, u.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token签发失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":    existingUser.ID,
			"name":  existingUser.Name,
			"email": existingUser.Email,
		},
	})
}

func Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	c.JSON(http.StatusOK, gin.H{"user_id": userID, "role": role})
}

func GetUserID(c *gin.Context) (uint, bool) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户id错误"})
		return 0, false
	}
	//类型断言
	userId, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息类型错误"})
		return 0, false
	}
	return userId, true
}

func GetParamID(c *gin.Context) (uint, bool) {
	idString := c.Param("id")
	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id错误"})
		return 0, false
	}
	return uint(id), true
}

func parsePage(c *gin.Context) (page, pageSize, offset int) {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err = strconv.Atoi(c.Query("page_size"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset = (page - 1) * pageSize
	return
}
