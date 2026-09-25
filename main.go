package main

import (
	"log"
	"paopao/internal/service"

	"paopao/internal/config"
	db2 "paopao/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("配置加载失败: ", err)
	}
	db, err := db2.LinkMySQL(cfg)
	if err != nil {
		log.Fatal("数据库连接失败", err)
	}
	err = db2.AutoMigrate(db)
	if err != nil {
		log.Fatal("数据库迁移失败: ", err)
	}
	r := service.NewRouter(db, cfg)
	_ = r.Run(":8080")
}
