package model

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// DB 数据库链接单例
var DB *gorm.DB

// Database 在中间件中初始化mysql链接
func Init() {
	dbhost := os.Getenv("db_host")
	dbport := os.Getenv("db_port")
	dbuser := os.Getenv("db_user")
	dbpassword := os.Getenv("db_password")
	dbname := os.Getenv("db_name")
	if dbport == "" {
		dbport = "3306"
	}
	dsn := dbuser + ":" + dbpassword + "@tcp(" + dbhost + ":" + dbport + ")/" + dbname + "?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open("mysql", dsn)
	db.LogMode(true)
	// Error
	if err != nil {
		panic(err)
	}
	if gin.Mode() == "release" {
		db.LogMode(false)
	}
	//设置连接池
	//空闲
	db.DB().SetMaxIdleConns(20)
	//打开
	db.DB().SetMaxOpenConns(100)
	//超时
	db.DB().SetConnMaxLifetime(time.Second * 30)

	DB = db

	//自动迁移
	//DB.Set("gorm:table_options", "charset=utf8mb4").
	//    AutoMigrate(&User{})
}
