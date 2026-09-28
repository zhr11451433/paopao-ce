package service

import (
	"errors"
	"net/http"
	"paopao/internal/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FollowHandler struct {
	db *gorm.DB
}

func NewFollowHandler(db *gorm.DB) *FollowHandler {
	return &FollowHandler{db: db}
}

//关注 /users/:id/follow

func (u *FollowHandler) Follow(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	followID, ok := GetParamID(c)
	if !ok {
		return
	}
	if userID == followID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "您无法关注自己"})
		return
	}
	var existingFollow db.User
	if err := u.db.Where("id = ?", followID).First(&existingFollow).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	var existingFollowUser db.Follow
	err := u.db.Where("user_id = ? AND follow_id = ?", userID, followID).First(&existingFollowUser).Error
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "您已关注该用户"})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	res := db.Follow{
		FollowID: followID,
		UserID:   userID,
	}
	if err := u.db.Create(&res).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "关注成功"})
}

// 取消关注 /users/:id/follow

func (u *FollowHandler) Unfollow(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	followID, ok := GetParamID(c)
	if !ok {
		return
	}

	result := u.db.Where("user_id = ? and follow_id = ?", userID, followID).Delete(&db.Follow{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "成功取消关注"})
}

//查粉丝/关注列表

type RespFollow struct {
	FollowID uint   `json:"follow_id"`
	Name     string `json:"name"`
}

//查关注

func (u *FollowHandler) FollowList(c *gin.Context) {
	userID, ok := GetParamID(c)
	if !ok {
		return
	}
	page, pageSize, offset := parsePage(c)
	var existingUser db.User
	if err := u.db.Where("id = ?", userID).First(&existingUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	var total int64
	if err := u.db.Model(&db.Follow{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	//我关注的人
	var follow []db.User
	if err := u.db.
		Select("users.id,users.name").
		Joins("JOIN follows on users.id = follows.follow_id and follows.user_id = ?", userID).
		Where("follows.user_id = ?", userID).
		//Preload("Follow").
		Limit(pageSize).
		Offset(offset).
		Order("follows.created_at desc").
		Find(&follow).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	res := make([]RespFollow, 0, len(follow))
	for _, v := range follow {
		res = append(res, RespFollow{
			FollowID: v.ID,
			Name:     v.Name,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"user":      existingUser.Name,
		"total":     total,
		"page_size": pageSize,
		"page":      page,
		"follows":   res,
	})
}

//查粉丝

type RespFan struct {
	FanID uint   `json:"fan_id"`
	Name  string `json:"name"`
}

func (u *FollowHandler) FanList(c *gin.Context) {
	userID, ok := GetParamID(c)
	if !ok {
		return
	}
	var existingUser db.User
	if err := u.db.Where("id = ?", userID).First(&existingUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	page, pageSize, offset := parsePage(c)
	var total int64
	if err := u.db.Model(&db.Follow{}).Where("follow_id = ?", userID).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	//关注我的人
	var fan []db.User
	if err := u.db.
		Select("users.id,users.name").
		Joins("JOIN follows on users.id = follows.user_id and follows.follow_id = ?", userID).
		Where("follows.follow_id = ?", userID).
		//Preload("Follow").
		Limit(pageSize).
		Offset(offset).
		Order("follows.created_at desc").
		Find(&fan).
		Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	res := make([]RespFan, 0, len(fan))
	for _, v := range fan {
		res = append(res, RespFan{
			FanID: v.ID,
			Name:  v.Name,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"user":      existingUser.Name,
		"total":     total,
		"page_size": pageSize,
		"page":      page,
		"fans":      res,
	})
}
