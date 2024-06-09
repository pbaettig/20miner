package comments

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/pbaettig/20miner/internal/config"
	"github.com/pbaettig/20miner/internal/pkg/utils"
	"gorm.io/gorm"
)

type Reactions struct {
	gorm.Model

	CommentID uint

	Awesome     int `json:"awesome"`
	Bad         int `json:"bad"`
	Nonsense    int `json:"nonsense"`
	Unnecessary int `json:"unnecessary"`
	Smart       int `json:"smart"`
	Exact       int `json:"exact"`
}

type Comment struct {
	ID        uint `json:"-"`
	ArticleID uint `json:"-"`

	Parent     *Comment `json:"-"`
	ParentID   *uint    `json:"-"`
	OriginalID string   `json:"id"`

	AuthorNickname string    `json:"authorNickname"`
	Body           string    `json:"body"`
	CounterSpeech  bool      `json:"counterSpeech"`
	CreatedAt      time.Time `json:"createdAt"`
	Reactions      Reactions `json:"reactions"`
}

// commentRepliesNested represents a Comment with all it's replies
// nested under the `Replies` property
type commentRepliesNested struct {
	Comment
	Replies []Comment `json:"replies"`
}

// FlattenReplies turns a Comment with nested Replies into
// a list of Comments, where replies are now comments with the
// ParentID set
func (c commentRepliesNested) FlattenReplies() []*Comment {
	replies := make([]*Comment, 0)

	cc := &c.Comment
	cc.SetID()
	cc.ParentID = nil

	// Parent comment
	replies = append(replies, cc)

	for _, reply := range c.Replies {
		reply.ParentID = &cc.ID
		replies = append(replies, &reply)

	}

	return replies
}

// SetID generates a uint ID from the original alphanumeric
// ID used on the website
func (c *Comment) SetID() {
	c.ID = utils.StringToUintHash(c.OriginalID)
}

type commentsRepliesNested []commentRepliesNested

// Flatten all comments with nested replies to a single list
// of comments with the ParentID set
// same as commentsRepliesNested.Flatten but for a collection of
// commentRepliesNested
func (cs commentsRepliesNested) Flatten() []*Comment {
	flattened := make([]*Comment, 0)
	for _, c := range cs {
		flattened = append(flattened, c.FlattenReplies()...)
	}

	return flattened
}

type commentApiResponse struct {
	CommentingEnabled bool                   `json:"commentingEnabled"`
	NextLink          string                 `json:"nextLink"`
	TotalCount        int                    `json:"totalCount"`
	Comments          []commentRepliesNested `json:"comments"`
}

// getAllComments gets all comments for a given article
func getAllComments(articleID string) commentsRepliesNested {
	params := url.Values{
		"tenantId":  []string{"6"},
		"contentId": []string{articleID},
		"limit":     []string{"50"},
	}

	u, err := url.Parse(config.CommentsAPIURL)
	if err != nil {
		log.Fatal(err)
	}

	u.RawQuery = params.Encode()
	uri := u.String()
	allComments := make(commentsRepliesNested, 0)

	for {
		resp := getCommentsFromUri(uri)

		for i := range resp.Comments {
			// generate uint primary key for GORM
			resp.Comments[i].ID = utils.StringToUintHash(resp.Comments[i].OriginalID)

			// Convert Article ID to uint so it can be used as a proper foreign key
			resp.Comments[i].ArticleID = utils.MustIntToUint(articleID)
		}

		allComments = append(allComments, resp.Comments...)

		if resp.NextLink == "" {
			break
		}

		uri = resp.NextLink
	}

	return allComments

}

func getCommentsFromUri(uri string) commentApiResponse {
	resp, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	cr := commentApiResponse{}

	if err = json.Unmarshal(buf, &cr); err != nil {
		log.Fatal(err)
	}

	return cr
}

func GetComments(articleID string) []*Comment {
	csrs := commentsRepliesNested(getAllComments(articleID))
	flattened := csrs.Flatten()
	for _, c := range flattened {
		c.SetID()
	}
	return flattened
}
