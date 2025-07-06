package internal

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

var db *gorm.DB

func DB() *gorm.DB {
	return db
}

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err, "db.go, init()")
	}
	db.AutoMigrate(&Script{}, &Role{}, &Sentence{})
}

type Script struct {
	Id      int64  `gorm:"primaryKey"`
	Sid     string `gorm:"uniqueIndex"`
	Title   string `gorm:"not null;type:varchar(255)"`
	Content string `gorm:"not null:type:text"`
}

type Role struct {
	Id            int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Rid           string `json:"rid"`
	Sid           string `json:"sid"`
	Name          string `json:"name"`
	Character     string `json:"character"`
	LanguageHabit string `json:"language_habit"`
}

type Sentence struct {
	Id              int64  `gorm:"primaryKey;autoIncrement"`
	RoleIdAssistant string `gorm:"role_id_assistant"`
	RoleIdUser      string `gorm:"role_id_user"`
	Role            string `gorm:"role"`
	Content         string `gorm:"content"`
}
