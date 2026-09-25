package db

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string `gorm:"size:32"`
	Email    string `gorm:"uniqueIndex;size:255"`
	Password string `json:"-"`
	Role     string `gorm:"size:16"`
}

//用户的帖子

type Post struct {
	gorm.Model
	UserID   uint      `gorm:"index"`
	User     User      `gorm:"foreignKey:UserID"`
	Title    string    `gorm:"type:varchar(100)"`
	Content  string    `gorm:"type:text"`
	Comments []Comment `gorm:"foreignKey:PostID"`
	//LikeCount int64     //点赞总数
	//收藏数
	//Like bool `gorm:"-"`//帖子是否点赞
}

//评论 没有标题

type Comment struct {
	gorm.Model
	UserID  uint   `gorm:"index"`
	User    User   `gorm:"foreignKey:UserID"`
	Content string `gorm:"type:varchar(512)"`
	PostID  uint   `gorm:"index"`
	//LikeCount int64  //评论点赞数
	//Like bool `gorm:"-"`//评论是否点赞
}

type PostLike struct {
	ID        uint `gorm:"primaryKey"`
	PostID    uint `gorm:"uniqueIndex:idx_post_user"` //赞了哪条帖子
	UserID    uint `gorm:"uniqueIndex:idx_post_user"` //谁点的
	CreatedAt time.Time
	//两个字段都用 uniqueIndex:同一个名字，GORM 就会把它们合成一个联合唯一索引。
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Post{},
		&Comment{},
		&PostLike{},
	)
}
