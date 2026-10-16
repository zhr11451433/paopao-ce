package service

import (
	"errors"
	"net/http"
	"paopao/internal/db"
	"sort"
	"strings"
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
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Author    string    `json:"author"`
	UserID    uint      `json:"user_id"`    // ← 新增：作者 id，前端拿它跳主页
	LikeCount int       `json:"like_count"` // ← 新增
	IsLiked   bool      `json:"is_liked"`   // ← 新增：当前用户点没点过
	// 当前用户是否关注了作者。游客、以及看自己的帖子，都是 false
	IsFollowing bool          `json:"is_following"`
	Comments    []CommentItem `json:"comments"`
	Images      []ImageItem   `json:"images"`
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
	UserID       uint        `json:"user_id"` // ← 新增
	CommentCount int         `json:"comment_count"`
	LikeCount    int         `json:"like_count"` // ← 新增
	IsLiked      bool        `json:"is_liked"`   // ← 新增
	IsFollowing  bool        `json:"is_following"`
	Images       []ImageItem `json:"images"`
}

type PostListItemMine struct {
	ID           uint        `json:"id"`
	Title        string      `json:"title"`
	CreatedAt    time.Time   `json:"created_at"`
	UserID       uint        `json:"user_id"`
	CommentCount int         `json:"comment_count"`
	LikeCount    int         `json:"like_count"`
	IsLiked      bool        `json:"is_liked"`
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
		Select("id", "title", "created_at", "user_id").
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
	likeMap, err := likeCountMap(u.db, postIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrDatabase})
		return
	}
	likedSet, err := likedPostSet(u.db, userId, postIDs)
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
			UserID:       post.UserID,
			CommentCount: countMap[post.ID],
			LikeCount:    likeMap[post.ID],
			IsLiked:      likedSet[post.ID],
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

// escapeLike 把用户输入里的 \ % _ 转成字面量。
// MySQL 默认用 \ 做 LIKE 的转义符，不转义的话搜一个 % 会匹配所有帖子。
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// postsByQuery 是广场列表的查询。q 为空就是全部帖子。
// 作者 Home.vue:347-350 把地址上的搜索词交给 getPosts；这里按标题或正文包含匹配，不接搜索引擎。
// 话题（type=tag）不是这一轮的事。热门排序在 List 里按 style 加。
func postsByQuery(dbConn *gorm.DB, q string) *gorm.DB {
	tx := dbConn.Model(&db.Post{})
	q = strings.TrimSpace(q)
	if q == "" {
		return tx
	}
	pattern := "%" + escapeLike(q) + "%"
	return tx.Where("title LIKE ? OR content LIKE ?", pattern, pattern)
}

//公开列表

func (u *PostHandler) List(c *gin.Context) {
	// ① 先拿"我是谁"。游客是 0 —— 列表允许游客看，只是 is_liked 恒为 false。
	myID := optionalUserID(c)

	page, pageSize, offset := parsePage(c)
	// Count 和 Find 各开一条查询。复用同一个 GORM 链的话，Count 会把 SELECT 改掉，下一句 Find 就乱了。
	q := c.Query("q")
	var total int64
	if err := postsByQuery(u.db, q).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	var posts []db.Post
	listQuery := postsByQuery(u.db, q).
		Preload("User").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort asc")
		})
	// 作者 Home.vue 的热门推荐把 style=hots 交给同一个列表。
	// 点赞记在 post_likes 里，帖子表没有点赞数列，所以用子查询计数。
	// 没人赞过的是 0，仍然出现。赞数相同再按 id 倒序，翻页才不会乱。
	// 其它 style（包括不传）保持原来的新帖在前。正在关注走的是 GET /feed，不在这里。
	if c.Query("style") == "hots" {
		listQuery = listQuery.
			Order("(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id) DESC").
			Order("posts.id DESC")
	} else {
		listQuery = listQuery.Order("id DESC")
	}
	if err := listQuery.
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
	// ② 三个批量查询，各一次。不要挪进下面的 for 循环里逐条查。
	countMap, err := commentCountMap(u.db, postIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	likeMap, err := likeCountMap(u.db, postIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	likedSet, err := likedPostSet(u.db, myID, postIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	authorIDs := make([]uint, 0, len(posts))
	for _, p := range posts {
		authorIDs = append(authorIDs, p.UserID)
	}
	followingSet, err := followingAuthorSet(u.db, myID, authorIDs)
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
			UserID:       p.UserID, // ← 新增
			CommentCount: countMap[p.ID],
			LikeCount:    likeMap[p.ID],  // ← 新增，没点过的是 0
			IsLiked:      likedSet[p.ID], // ← 新增，没点过的是 false
			IsFollowing:  followingSet[p.UserID],
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

// ListByUser 是某人发的帖子：GET /users/:id/posts。
//
// 作者 User.vue 的 loadPostsByStyle 还会按 style 分成评论 / 亮点 / 图文 / 喜欢，
// 那是 C55，这里只出他自己写的帖子。
// 每一项的组装方式和上面的 List 相同，只是多了 user_id 条件。
// 公开接口：游客也能看，is_liked 在未登录时是 false。
func (u *PostHandler) ListByUser(c *gin.Context) {
	userID, ok := GetParamID(c)
	if !ok {
		return
	}
	myID := optionalUserID(c)

	var existing db.User
	if err := u.db.First(&existing, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	page, pageSize, offset := parsePage(c)
	var total int64
	if err := u.db.Model(&db.Post{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	var posts []db.Post
	if err := u.db.
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort asc")
		}).
		Where("user_id = ?", userID).
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
	likeMap, err := likeCountMap(u.db, postIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	likedSet, err := likedPostSet(u.db, myID, postIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	followingSet, err := followingAuthorSet(u.db, myID, []uint{userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	list := make([]PostListItem, 0, len(posts))
	for _, p := range posts {
		list = append(list, PostListItem{
			ID:           p.ID,
			Title:        p.Title,
			CreatedAt:    p.CreatedAt,
			Author:       existing.Name,
			UserID:       p.UserID,
			CommentCount: countMap[p.ID],
			LikeCount:    likeMap[p.ID],
			IsLiked:      likedSet[p.ID],
			IsFollowing:  followingSet[p.UserID],
			Images:       toImageItems(p.Images),
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
	myID := optionalUserID(c)

	var post db.Post
	if err := u.db.
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort asc")
		}).
		Preload("User").
		Preload("Comments.User").
		First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	// 点赞数：单条帖子，直接 Count
	var likeCount int64
	if err := u.db.Model(&db.PostLike{}).
		Where("post_id = ?", post.ID).
		Count(&likeCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}
	// 我点没点过：游客（myID == 0）不用查，恒 false
	isLiked := false
	if myID != 0 {
		var n int64
		if err := u.db.Model(&db.PostLike{}).
			Where("user_id = ? AND post_id = ?", myID, post.ID).
			Count(&n).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
			return
		}
		isLiked = n > 0
	}
	followingSet, err := followingAuthorSet(u.db, myID, []uint{post.UserID})
	if err != nil {
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
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})
	images := toImageItems(post.Images)
	c.JSON(http.StatusOK, PostDetail{
		ID:          post.ID,
		Title:       post.Title,
		Content:     post.Content,
		CreatedAt:   post.CreatedAt,
		Author:      post.User.Name,
		UserID:      post.UserID,    // ← 新增
		LikeCount:   int(likeCount), // ← 新增
		IsLiked:     isLiked,        // ← 新增
		IsFollowing: followingSet[post.UserID],
		Comments:    comments,
		Images:      images,
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

// 批量查一组帖子的点赞数。和 commentCountMap 一个套路：一次查完，按 post_id 分组。
// 没被点赞过的帖子不会出现在 map 里，取值时用 likeMap[id] 拿到的是 0（map 的零值）。
func likeCountMap(u *gorm.DB, postIDs []uint) (map[uint]int, error) {
	counts := make(map[uint]int, len(postIDs))
	if len(postIDs) == 0 {
		return counts, nil
	}
	type row struct {
		PostID uint
		Count  int
	}
	var rows []row
	if err := u.Model(db.PostLike{}).
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

// 批量查「这一组帖子里，我点过哪几个」。
// myID == 0（游客）时直接返回空 map，一次数据库都不查。
func likedPostSet(u *gorm.DB, myID uint, postIDs []uint) (map[uint]bool, error) {
	liked := make(map[uint]bool, len(postIDs))
	if myID == 0 || len(postIDs) == 0 {
		return liked, nil
	}
	var ids []uint
	if err := u.Model(&db.PostLike{}).
		Where("user_id = ? AND post_id IN ?", myID, postIDs).
		Pluck("post_id", &ids).Error; err != nil {
		return nil, err
	}
	for _, id := range ids {
		liked[id] = true
	}
	return liked, nil
}

// followingAuthorSet 当前用户关注了这批作者里的哪些人。
// follows.user_id 是谁，follow_id 是被关注的人。
// myID 为 0（未登录）不查库，取值是 false。看自己的帖子也不会出现在结果里。
func followingAuthorSet(u *gorm.DB, myID uint, authorIDs []uint) (map[uint]bool, error) {
	set := make(map[uint]bool)
	if myID == 0 || len(authorIDs) == 0 {
		return set, nil
	}
	var ids []uint
	if err := u.Model(&db.Follow{}).
		Where("user_id = ? AND follow_id IN ?", myID, authorIDs).
		Pluck("follow_id", &ids).Error; err != nil {
		return nil, err
	}
	for _, id := range ids {
		set[id] = true
	}
	return set, nil
}
