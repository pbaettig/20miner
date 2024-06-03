package main

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/pbaettig/20miner/internal/config"
	"github.com/pbaettig/20miner/internal/pkg/articles"
	"github.com/pbaettig/20miner/internal/pkg/comments"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func main() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
	})

	// Migrate the schema
	db.AutoMigrate(&articles.Article{})
	db.AutoMigrate(&articles.Shares{})
	db.AutoMigrate(&comments.Comment{})
	db.AutoMigrate(&comments.Reactions{})

	log.Infof("getting front page...")
	articleLinks, err := articles.GetArticleLinks(client, config.FrontURL)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("got %d article links from front page", len(articleLinks))

	// articles := make([]articles.Article, 0)
	for i, al := range articleLinks {
		fields := log.Fields{
			"num":  i + 1,
			"href": al.Href,
		}
		log.WithFields(fields).Info("grabbing article")
		article := al.Get(client)

		article.Comments = comments.GetComments(article.OriginalID)

		// articles = append(articles, article)
		log.WithFields(fields).Info("inserting to DB")
		if tx := db.Create(&article); tx.Error != nil {
			log.Error(tx.Error)
		}

	}

	// buf, err := json.Marshal(articles)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fd, err := os.OpenFile("articles.json", os.O_CREATE|os.O_RDWR, 0755)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer fd.Close()

	// if _, err := fd.Write(buf); err != nil {
	// 	log.Fatal(err)
	// }
}
