package service

import (
	"errors"
	"net/http"
	"paopao/internal/db"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommentHandler struct {
	db *gorm.DB
}

func NewCommentHandler(db *gorm.DB) *CommentHandler {
	return &CommentHandler{db: db}
}

//方法	        路径	         登录  干什么
//POST /posts/:id/comments 是 对这篇帖子评论
//GET  /posts/:id/comments 否 这篇的评论列表（可分页）
//DELETE /comments/:id     是  删自己的评论
//发表评论

type CommentReq struct {
	Content string `json:"content" binding:"required,max=512"`
}

func (u *CommentHandler) Post(c *gin.Context) {
	postID, ok := GetParamID(c)
	if !ok {
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	var comment CommentReq
	err := c.ShouldBindJSON(&comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析失败"})
		return
	}
	var existingPost db.Post
	if err := u.db.Where("id=?", postID).First(&existingPost).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	saveComment := db.Comment{
		UserID:  userID,
		Content: comment.Content,
		PostID:  postID,
	}
	if err := u.db.Create(&saveComment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "评论失败"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "评论成功"})
}

//取这篇帖子的评论列表

type CommentList struct {
	ID        uint      `json:"id"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	Content   string    `json:"content"`
}

func (u *CommentHandler) Get(c *gin.Context) {
	postID, ok := GetParamID(c)
	if !ok {
		return
	}
	page, pageSize, offset := parsePage(c)
	var existingPost db.Post
	if err := u.db.Where("id=?", postID).First(&existingPost).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	//帖子存在 查评论
	var comments []db.Comment
	if err := u.db.Preload("User").
		Where("post_id=?", postID).
		Order("created_at asc").
		Limit(pageSize).
		Offset(offset).
		Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	//评论总数
	var total int64
	if err := u.db.Model(&db.Comment{}).Where("post_id = ?", postID).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	commentLists := make([]CommentList, 0, len(comments))
	for _, comment := range comments {
		commentLists = append(commentLists, CommentList{
			ID:        comment.ID,
			Author:    comment.User.Name,
			CreatedAt: comment.CreatedAt,
			Content:   comment.Content,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"comments":  commentLists,
	})
}

//删评论

func (u *CommentHandler) findOwnComment(c *gin.Context, userId, id uint) (*db.Comment, bool) {
	var comment db.Comment
	if err := u.db.First(&comment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "评论不存在"})
			return nil, false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return nil, false
	}
	if comment.UserID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作"})
		return nil, false
	}
	return &comment, true
}

func (u *CommentHandler) Delete(c *gin.Context) {
	commentID, ok := GetParamID(c)
	if !ok {
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	comment, ok := u.findOwnComment(c, userID, commentID)
	if !ok {
		return
	}
	if err := u.db.Delete(comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "删除成功"})
}
