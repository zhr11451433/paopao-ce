package service

import (
	"net/http"
	"paopao/internal/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

//feed流

type FeedHandler struct {
	db *gorm.DB
}

func NewFeedHandler(db *gorm.DB) *FeedHandler {
	return &FeedHandler{
		db: db,
	}
}

type FollowID struct {
	FollowID uint `json:"follow_id"`
}

func (u *FeedHandler) Feed(c *gin.Context) {
	myId, ok := GetUserID(c)
	if !ok {
		return
	}
	var FollowIDs []FollowID
	if err := u.db.Model(&db.Follow{}).
		Where("user_id = ?", myId).
		Find(&FollowIDs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	followIds := make([]uint, 0, len(FollowIDs))
	for _, d := range FollowIDs {
		followIds = append(followIds, d.FollowID)
	}
	page, pageSize, offset := parsePage(c)
	var Posts []db.Post
	if err := u.db.
		Preload("User").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort acs")
		}).
		Where("user_id IN (?) or user_id = ?", followIds, myId).
		Order("posts.id desc").
		Limit(pageSize).
		Offset(offset).
		Find(&Posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	var total int64
	if err := u.db.Model(&db.Post{}).Where("user_id IN (?) or user_id = ?", followIds, myId).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	postIDs := make([]uint, 0, len(Posts))
	for _, post := range Posts {
		postIDs = append(postIDs, post.ID)
	}
	countMap, err := commentCountMap(u.db, postIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	requestPosts := make([]PostListItem, 0, len(Posts))
	for _, post := range Posts {
		// ① 先把这篇帖子的图片转成 DTO（同样是"长度 0 + append"）
		images := toImageItems(post.Images)
		requestPosts = append(requestPosts, PostListItem{
			ID:           post.ID,
			Author:       post.User.Name,
			CreatedAt:    post.CreatedAt,
			Title:        post.Title,
			CommentCount: countMap[post.ID],
			Images:       images,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"posts":     requestPosts,
	})
}
