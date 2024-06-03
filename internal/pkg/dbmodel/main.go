package dbmodel

import (
	"time"

	"github.com/pbaettig/20miner/internal/pkg/articles"
	"github.com/pbaettig/20miner/internal/pkg/comments"
)

type ArticleRow struct {
	articles.Article
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time    `gorm:"index"`
	Comments  []CommentRow `gorm:"foreignKey:ArticleID"`
}

func (a *ArticleRow) SetComments(cs []comments.Comment) {
	a.Comments = make([]CommentRow, 0)
	for _, c := range cs {
		a.Comments = append(a.Comments, CommentRow{Comment: c})
	}
}

type CommentRow struct {
	comments.Comment
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time `gorm:"index"`
}
