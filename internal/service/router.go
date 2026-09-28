package service

import (
	"paopao/internal/config"
	"paopao/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	Auth := middleware.Auth(cfg.JWTSecret)
	//Require := middleware.RequireAdmin()
	userHandler := NewUserHandler(db, cfg.JWTSecret)
	postHandler := NewPostHandler(db)
	commentHandler := NewCommentHandler(db)
	likeHandler := NewLikeHandler(db)
	followHandler := NewFollowHandler(db)
	r.GET("/me", Auth, Me)
	r.POST("/register", userHandler.Register)
	r.POST("/login", userHandler.Login)
	//帖子
	post := r.Group("/posts")
	{
		post.GET("/me", Auth, postHandler.ListMine)
		post.POST("", Auth, postHandler.Create)
		post.PUT("/:id", Auth, postHandler.Update)
		post.DELETE("/:id", Auth, postHandler.Delete)
		post.GET("/:id", postHandler.GetOne)
		post.GET("", postHandler.List)
		//评论
		post.POST("/:id/comments", Auth, commentHandler.Post)
		post.GET("/:id/comments", commentHandler.Get)
		//点赞
		post.POST("/:id/likes", Auth, likeHandler.Like)
		post.DELETE("/:id/likes", Auth, likeHandler.Delete)
		post.GET("/me/likes", Auth, likeHandler.GetMe)
		post.GET("/:id/likes", likeHandler.GetOne)
	}
	r.DELETE("/comments/:id", Auth, commentHandler.Delete)
	user := r.Group("/users")
	{
		//关注列表
		user.GET("/:id/following", followHandler.FollowList)
		user.GET("/:id/followers", followHandler.FanList)
		user.POST("/:id/follow", Auth, followHandler.Follow)
		user.DELETE("/:id/follow", Auth, followHandler.Unfollow)
	}
	return r
}
