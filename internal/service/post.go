package service

import (
	"errors"
	"net/http"
	"paopao/internal/db"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 业务哨兵错误

var (
	ErrImageCountOverflow = errors.New("一次最多上传9张图片")
	ErrImageInvalid       = errors.New("部分图片无效：图片不存在、不属于你或已绑定其他帖子")
	ErrDatabase           = errors.New("数据库错误")
)

type PostHandler struct {
	db *gorm.DB
}

func NewPostHandler(db *gorm.DB) *PostHandler {
	return &PostHandler{db: db}
}

type PostRequest struct {
	Title   string `json:"title" binding:"required,max=100"`
	Content string `json:"content" binding:"required"`
	Images  []uint `json:"images"` // 新增：图片id数组，可选
}

type CommentItem struct {
	ID        uint      `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

type PostDetail struct {
	ID        uint          `json:"id"`
	Title     string        `json:"title"`
	Content   string        `json:"content"`
	CreatedAt time.Time     `json:"created_at"`
	Author    string        `json:"author"`
	Comments  []CommentItem `json:"comments"`
	Images    []ImageItem   `json:"images"`
}

//type PostList struct {
//	ID        uint        `json:"id"`
//	Title     string      `json:"title"`
//	CreatedAt time.Time   `json:"created_at"`
//	Images    []ImageItem `json:"images"`
//}

type PostListItem struct {
	ID           uint        `json:"id"`
	Title        string      `json:"title"`
	CreatedAt    time.Time   `json:"created_at"`
	Author       string      `json:"author"`
	CommentCount int         `json:"comment_count"`
	Images       []ImageItem `json:"images"`
}

type PostListItemMine struct {
	ID           uint        `json:"id"`
	Title        string      `json:"title"`
	CreatedAt    time.Time   `json:"created_at"`
	CommentCount int         `json:"comment_count"`
	Images       []ImageItem `json:"images"`
}

func (u *PostHandler) findOwnPost(c *gin.Context, userId, id uint) (*db.Post, bool) {
	var post db.Post
	if err := u.db.First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
			return nil, false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return nil, false
	}
	if post.UserID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作"})
		return nil, false
	}
	return &post, true
}

func commentCountMap(u *gorm.DB, postIDs []uint) (map[uint]int, error) {
	counts := make(map[uint]int, len(postIDs))
	if len(postIDs) == 0 {
		return counts, nil
	}
	type row struct {
		PostID uint
		Count  int
	}
	var rows []row
	if err := u.Model(db.Comment{}).
		Select("post_id, COUNT(*) as count").
		Where("post_id IN ?", postIDs).
		Group("post_id").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		counts[r.PostID] = r.Count
	}
	return counts, nil
}

func (u *PostHandler) ListMine(c *gin.Context) {
	userId, ok := GetUserID(c)
	if !ok {
		return
	}
	page, pageSize, offset := parsePage(c)
	var total int64
	if err := u.db.Model(&db.Post{}).Where("user_id = ?", userId).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrDatabase})
		return
	}
	var posts []db.Post
	if err := u.db.
		Select("id", "title", "created_at").
		Where("user_id = ?", userId).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort asc")
		}).
		Order("id DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrDatabase})
		return
	}
	list := make([]PostListItemMine, 0, len(posts))
	postIDs := make([]uint, 0, len(posts))
	for _, p := range posts {
		postIDs = append(postIDs, p.ID)
	}
	countMap, err := commentCountMap(u.db, postIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrDatabase})
		return
	}
	for _, post := range posts {
		images := toImageItems(post.Images)
		list = append(list, PostListItemMine{
			ID:           post.ID,
			Title:        post.Title,
			CreatedAt:    post.CreatedAt,
			CommentCount: countMap[post.ID],
			Images:       images,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"list":      list,
	})
}

//公开列表

func (u *PostHandler) List(c *gin.Context) {
	page, pageSize, offset := parsePage(c)
	var total int64
	if err := u.db.Model(&db.Post{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	var posts []db.Post
	if err := u.db.
		Preload("User").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort asc")
		}).
		Order("id DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	postIDs := make([]uint, 0, len(posts))
	for _, p := range posts {
		postIDs = append(postIDs, p.ID)
	}
	countMap, err := commentCountMap(u.db, postIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	list := make([]PostListItem, 0, len(posts))
	for _, p := range posts {
		images := toImageItems(p.Images)
		list = append(list, PostListItem{
			ID:           p.ID,
			Title:        p.Title,
			CreatedAt:    p.CreatedAt,
			Author:       p.User.Name,
			CommentCount: countMap[p.ID],
			Images:       images,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"list":      list,
	})
}

//发帖

//⚠️ 已被 PostImageHandler.Posts 取代，勿挂路由」
//func (u *PostHandler) Create(c *gin.Context) {
//	userId, ok := GetUserID(c)
//	if !ok {
//		return
//	}
//	var req PostRequest
//	if err := c.ShouldBindJSON(&req); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "解析失败"})
//		return
//	}
//	savePost := db.Post{
//		UserID:  userId,
//		Title:   req.Title,
//		Content: req.Content,
//	}
//	if err := u.db.Create(&savePost).Error; err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
//		return
//	}
//	c.JSON(http.StatusCreated, gin.H{
//		"status": "创建成功",
//		"id":     savePost.ID,
//	})
//}

//改帖

func (u *PostHandler) Update(c *gin.Context) {
	userId, ok := GetUserID(c)
	if !ok {
		return
	}
	id, ok := GetParamID(c)
	if !ok {
		return
	}
	var req PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析失败"})
		return
	}
	post, ok := u.findOwnPost(c, userId, id)
	if !ok {
		return
	}
	if len(req.Images) > 9 {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrImageCountOverflow})
		return
	}
	err := u.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(post).Updates(map[string]any{
			"title":   req.Title,
			"content": req.Content,
		}).Error; err != nil {
			return err
		}
		// ✅ 关键点：images字段没传，直接跳过全部图片处理，旧图片保持原样
		if req.Images == nil {
			return nil
		}
		// ② 将本帖子原来绑定的全部图片解除绑定：post_id置null，status变回0
		if err := tx.Model(&db.PostImage{}).
			Where("post_id = ?", post.ID).
			Updates(map[string]any{"post_id": nil, "status": 0}).Error; err != nil {
			return err
		}
		if len(req.Images) == 0 {
			return nil
		}
		// ③ 绑定新传入的图片ID列表
		// where条件：图片属于我，并且要么是临时未绑定，要么本来就属于这个帖子
		result := tx.Model(&db.PostImage{}).
			Where("id in ? and user_id = ? and (post_id is null or post_id = ?)", req.Images, userId, post.ID).
			Updates(map[string]any{
				"status":  1,
				"post_id": post.ID,
			})
		if result.Error != nil {
			return result.Error
		}
		// 行数校验：有任意一张图片不满足条件（不属于你 / 已经绑别的帖子 / id不存在），直接回滚
		if result.RowsAffected != int64(len(req.Images)) {
			return ErrImageInvalid
		}
		return nil
	})
	if err != nil {
		// 区分业务错误和数据库错误
		if errors.Is(err, ErrImageCountOverflow) || errors.Is(err, ErrImageInvalid) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "更新成功"})
	//if err := u.db.Model(post).Updates(map[string]any{
	//	"title":   req.Title,
	//	"content": req.Content,
	//}).Error; err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
	//	return
	//}
	//c.JSON(http.StatusOK, gin.H{"status": "更新成功"})
}

func (u *PostHandler) Delete(c *gin.Context) {
	userId, ok := GetUserID(c)
	if !ok {
		return
	}
	id, ok := GetParamID(c)
	if !ok {
		return
	}
	post, ok := u.findOwnPost(c, userId, id)
	if !ok {
		return
	}
	if err := u.db.Delete(post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "删除成功"})
}

//查单个帖子

type ImageItem struct {
	ID  uint   `json:"id"`
	Url string `json:"url"`
	//FileName string `json:"file_name"`
}

func (u *PostHandler) GetOne(c *gin.Context) {
	id, ok := GetParamID(c)
	if !ok {
		return
	}
	var post db.Post
	if err := u.db.
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort asc")
		}).
		Preload("User").
		Preload("Comments.User").
		//Order("sort asc").
		First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	comments := make([]CommentItem, 0, len(post.Comments))
	for _, cm := range post.Comments {
		comments = append(comments, CommentItem{
			ID:        cm.ID,
			Content:   cm.Content,
			Author:    cm.User.Name,
			CreatedAt: cm.CreatedAt,
		})
	}
	// ✅ 和评论接口保持一致 created_at ASC：时间小（更早）的排前面
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})
	images := toImageItems(post.Images)
	//sort.Slice(images, func(i, j int) bool {
	//	return images[i].ID < images[j].ID
	//})
	c.JSON(http.StatusOK, PostDetail{
		ID:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
		Author:    post.User.Name,
		Comments:  comments,
		Images:    images,
	})
}

//把 []db.PostImage 转成 []ImageItem

func toImageItems(images []db.PostImage) []ImageItem {
	items := make([]ImageItem, 0, len(images))
	for _, image := range images {
		items = append(items, ImageItem{
			ID:  image.ID,
			Url: image.Url,
		})
	}
	return items
}
