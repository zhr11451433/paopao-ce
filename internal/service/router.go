package service

import (
	"paopao/internal/config"
	"paopao/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8MB
	// 静态路由，访问 /uploads/xxx 映射到本地 ./uploads
	//
	// 【为什么必须有这行】图片上传成功后，数据库里存的是 url（如 /uploads/2026/02/18/xxx.jpg）。
	//   如果不用 Static 把它映射到磁盘目录，浏览器访问这个 url 会 404 ——
	//   文件明明在磁盘上，却谁也看不到。
	// 【另一个方案】r.StaticFS / r.StaticFile，或者生产环境直接交给 Nginx、
	//   对象存储（paopao-ce 就是走 OSS 接口，可切换 LocalOSS / AliOSS / MinIO）。
	// 【注意】路径改成用 uploadBaseDir 常量（定义在 post_image.go），
	//   保证"存盘的目录"和"对外暴露的目录"是同一个，不会哪天改了一处忘了另一处。
	//   它和上传一样是【相对于进程工作目录】的 —— 换个目录启动就会对不上。
	// ❌ 原代码：r.Static("/uploads", "./uploads")
	r.Static("/uploads", uploadBaseDir)

	Auth := middleware.Auth(cfg.JWTSecret)
	//Require := middleware.RequireAdmin()
	userHandler := NewUserHandler(db, cfg.JWTSecret)
	postHandler := NewPostHandler(db)
	commentHandler := NewCommentHandler(db)
	likeHandler := NewLikeHandler(db)
	followHandler := NewFollowHandler(db)
	feedHandler := NewFeedHandler(db)
	postImageHandler := NewPostImageHandler(db)
	r.GET("/me", Auth, Me)
	r.POST("/register", userHandler.Register)
	r.POST("/login", userHandler.Login)
	// 图片上传（根级，对应 post_image.go 里设计注释的 "① POST /upload"）
	r.POST("/upload", Auth, postImageHandler.Upload)
	//帖子
	post := r.Group("/posts")
	{
		post.GET("/me", Auth, postHandler.ListMine)
		post.POST("", Auth, postImageHandler.Posts)
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
		// ❌ 原代码把上传挂在 /posts 组里，实际路径是 /posts/upload：
		//    post.POST("/upload", Auth, postImageHandler.Upload)
		// 【问题 1】和你的设计注释不一致 —— 那里写的是 "① POST /upload"。
		// 【问题 2】语义不通：上传不是"帖子的子资源"，/posts/upload 容易被误读成
		//   "某个 id 叫 upload 的帖子"。而且它和 /posts/:id 挂在同一层，
		//   虽然 gin 里静态段优先、不会冲突，但读起来很别扭。
		// ✅ 挪到根级 → POST /upload（见下面 r.POST("/upload", ...)）
	}
	r.GET("/feed", Auth, feedHandler.Feed)
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
