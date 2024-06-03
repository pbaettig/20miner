package main

import (
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Article struct {
	ID    uint
	Title string
}

type ArticleRow struct {
	Article
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time    `gorm:"index"`
	Comments  []CommentRow `gorm:"foreignKey:ArticleID"`
}

type Comment struct {
	ID        uint
	ArticleID uint
	UserName  string
	Body      string
}

type CommentRow struct {
	Comment
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time `gorm:"index"`
}

func main() {
	db, err := gorm.Open(sqlite.Open("test2.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	cs := make([]CommentRow, 0)
	cs = append(cs, CommentRow{Comment: Comment{ID: 1, ArticleID: 1234, UserName: "pascal", Body: "yes"}})
	cs = append(cs, CommentRow{Comment: Comment{ID: 2, ArticleID: 1234, UserName: "pascal", Body: "no"}})
	cs = append(cs, CommentRow{Comment: Comment{ID: 3, ArticleID: 1234, UserName: "pascal", Body: "maybe?"}})

	a := &ArticleRow{
		Article:  Article{ID: 1234, Title: "Test 1234"},
		Comments: cs,
	}

	// Migrate the schema
	db.AutoMigrate(&ArticleRow{})
	db.AutoMigrate(&CommentRow{})

	// Create
	if tx := db.Create(a); tx.Error != nil {
		log.Fatal(tx.Error)
	}
}
