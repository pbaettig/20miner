package main

import (
	"flag"
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

var (
	dbPath      string
	httpTimeout time.Duration
)

func main() {
	flag.StringVar(&dbPath, "db", "", "path to the SQLite DB file")
	flag.DurationVar(&httpTimeout, "timeout", 5*time.Second, "timeout for any HTTP requests")
	flag.Parse()

	if dbPath == "" {
		log.Fatal("-db parameter is required")
	}

	client := &http.Client{
		Timeout: httpTimeout,
	}

	log.Infof("opening DB file %s", dbPath)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
	})

	log.Infof("performing schema auto-migration...")
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

	for i, al := range articleLinks {
		fields := log.Fields{
			"num":  i + 1,
			"href": al.Href,
		}
		log.WithFields(fields).Info("grabbing article")
		article := al.Get(client)

		article.Comments = comments.GetComments(article.OriginalID)

		log.WithFields(fields).Info("inserting to DB")
		if tx := db.Create(&article); tx.Error != nil {
			log.Error(tx.Error)
		}

	}
}
