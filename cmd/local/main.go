package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/pbaettig/20miner/internal/pkg/articles"
	"github.com/pbaettig/20miner/internal/pkg/comments"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func main() {
	as := make([]articles.Article, 0)

	fd, err := os.Open("articles.json")
	if err != nil {
		log.Fatal(err)
	}
	buf, err := io.ReadAll(fd)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(buf, &as); err != nil {
		log.Fatal(err)
	}

	fmt.Println(len(as))

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// db = db.Clauses(clause.OnConflict{
	// 	Columns:   []clause.Column{{Name: "id"}},
	// 	DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
	// })
	db = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"updated_at": time.Now()}),
	})

	// Migrate the schema
	db.AutoMigrate(&articles.Article{})
	db.AutoMigrate(&articles.Shares{})
	db.AutoMigrate(&comments.Comment{})
	db.AutoMigrate(&comments.Reactions{})

	for i, a := range as {

		fmt.Printf("#%d: inserting Article %s into DB...\n", i, a.OriginalID)

		if tx := db.Create(&a); tx.Error != nil {
			log.Println(tx.Error)
		}

		fmt.Println()
	}

}
