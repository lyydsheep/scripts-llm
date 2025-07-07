package internal

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

var db *gorm.DB

func DB() *gorm.DB {
	return db
}

func InitDB() {
	var err error
	db, err = gorm.Open(mysql.Open("root:root@tcp(mysql-container:3306)/dev?charset=utf8mb4&parseTime=True&loc=UTC"), &gorm.Config{})
	if err != nil {
		log.Fatal(err, "db.go, init()")
	}
	db.AutoMigrate(&Script{}, &Role{}, &Sentence{})
}

type Script struct {
	Id      int64  `gorm:"primaryKey"`
	Sid     string `gorm:"uniqueIndex;type:varchar(64)"`
	Title   string `gorm:"not null;type:varchar(255)"`
	Content string `gorm:"not null:type:text"`
}

type Role struct {
	Id            int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Rid           string `json:"rid" gorm:"type:varchar(64)"`
	Sid           string `json:"sid" gorm:"type:varchar(64)"`
	Name          string `json:"name" gorm:"type:varchar(64)"`
	Character     string `json:"character" gorm:"type:varchar(255)"`
	LanguageHabit string `json:"language_habit" gorm:"type:varchar(255)"`
}

type Sentence struct {
	Id              int64  `gorm:"primaryKey;autoIncrement"`
	RoleIdAssistant string `gorm:"column:role_id_assistant;type:varchar(64)"`
	RoleIdUser      string `gorm:"column:role_id_user;type:varchar(64)"`
	Role            string `gorm:"column:role;type:varchar(64)"`
	Content         string `gorm:"column:content;type:text"`
}
