package service

import (
	"errors"
	"net/http"
	"paopao/internal/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

//点赞

type LikeHandler struct {
	db *gorm.DB
}

func NewLikeHandler(db *gorm.DB) *LikeHandler {
	return &LikeHandler{db: db}
}

//点赞 POST	/posts/:id/like

func (u *LikeHandler) Like(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	postID, ok := GetParamID(c)
	if !ok {
		return
	}
	var existingPost db.Post
	if err := u.db.Where("id = ?", postID).First(&existingPost).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "点赞帖子不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	var existingPostLike db.PostLike
	if err := u.db.Where("user_id = ? AND post_id = ?", userID, postID).First(&existingPostLike).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
			return
		}
		//没找到，点赞成功
		postlike := db.PostLike{
			UserID: userID,
			PostID: postID,
		}
		if err := u.db.Create(&postlike).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "点赞失败"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "点赞成功"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "已点赞"})
}

//取消点赞 DELETE /posts/:id/like

func (u *LikeHandler) Delete(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	postID, ok := GetParamID(c)
	if !ok {
		return
	}
	result := u.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(db.PostLike{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	// else if .RowsAffected == 0 || result.RowsAffected > 0 {
	//	c.JSON(http.StatusOK, gin.H{"status": "已取消点赞"})
	//	return
	//}
	c.JSON(http.StatusOK, gin.H{"status": "已取消点赞"})
}

//查一个帖子的点赞数

func (u *LikeHandler) GetOne(c *gin.Context) {
	var likeCount int64
	postID, ok := GetParamID(c)
	if !ok {
		return
	}
	var post db.Post
	if err := u.db.Where("id = ?", postID).First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	if err := u.db.Model(&db.PostLike{}).Where("post_id = ?", postID).Count(&likeCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":         postID,
		"title":      post.Title,
		"like_count": likeCount,
	})
}

//查用户自己点赞了哪些帖子

type UserOwnLike struct {
	Author string `json:"author"`
	Title  string `json:"title"`
	ID     uint   `json:"id"`
}

func (u *LikeHandler) GetMe(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	var likePost []db.PostLike
	if err := u.db.Where("user_id = ?", userID).Find(&likePost).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	postID := make([]uint, 0, len(likePost))
	for _, like := range likePost {
		postID = append(postID, like.PostID)
	}
	if len(postID) == 0 {
		c.JSON(http.StatusOK, []UserOwnLike{})
		return
	}
	var likePosts []db.Post
	if err := u.db.
		Joins("JOIN post_likes ON post_likes.post_id = posts.id AND post_likes.user_id = ?", userID).
		Where("posts.id in ?", postID).
		Order("post_likes.created_at DESC").
		Preload("User").
		Find(&likePosts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	userOwnPosts := make([]UserOwnLike, 0, len(likePosts))
	for _, post := range likePosts {
		userOwnPosts = append(userOwnPosts, UserOwnLike{
			Author: post.User.Name,
			Title:  post.Title,
			ID:     post.ID,
		})
	}
	c.JSON(http.StatusOK, userOwnPosts)
}
