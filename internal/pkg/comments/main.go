package comments

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/pbaettig/20miner/internal/config"
)

type reactions struct {
	Awesome     int `json:"awesome"`
	Bad         int `json:"bad"`
	Nonsense    int `json:"nonsense"`
	Unnecessary int `json:"unnecessary"`
	Smart       int `json:"smart"`
	Exact       int `json:"exact"`
}

type commentBase struct {
	ID             string `json:"id"`
	ArticleID      string `json:"-"`
	AuthorNickname string `json:"authorNickname"`
	AuthorAvatar   struct {
		Light string `json:"light"`
		Dark  string `json:"dark"`
	} `json:"authorAvatar"`
	Body          string        `json:"body"`
	CounterSpeech bool          `json:"counterSpeech"`
	CreatedAt     time.Time     `json:"createdAt"`
	Reactions     reactions     `json:"reactions"`
	Notes         []interface{} `json:"notes"`
}

type commentRepliesNested struct {
	commentBase
	Replies []commentBase `json:"replies"`
}

func (c commentRepliesNested) FlattenReplies() []Comment {
	replies := make([]Comment, 0)

	// Parent comment
	replies = append(replies, Comment{
		commentBase: c.commentBase,
		ParentID:    "",
	})

	for _, reply := range c.Replies {
		replies = append(replies, Comment{commentBase: reply, ParentID: c.ID})
	}

	return replies
}

type commentsRepliesNested []commentRepliesNested

func (cs commentsRepliesNested) Flatten() []Comment {
	flattened := make([]Comment, 0)
	for _, c := range cs {
		flattened = append(flattened, c.FlattenReplies()...)
	}

	return flattened
}

type Comment struct {
	commentBase
	ArticleID string
	ParentID  string
}

type CommentApiResponse struct {
	CommentingEnabled bool                   `json:"commentingEnabled"`
	NextLink          string                 `json:"nextLink"`
	TotalCount        int                    `json:"totalCount"`
	Comments          []commentRepliesNested `json:"comments"`
}

func getAllComments(id string) commentsRepliesNested {
	params := url.Values{
		"tenantId":  []string{"6"},
		"contentId": []string{id},
		"limit":     []string{"50"},
		// "sortBy":    []string{"created_at"},
		// "sortOrder": []string{"desc"},
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
			resp.Comments[i].ArticleID = id
		}

		allComments = append(allComments, resp.Comments...)

		if resp.NextLink == "" {
			break
		}

		uri = resp.NextLink
	}

	return allComments

}

func getCommentsFromUri(uri string) CommentApiResponse {
	resp, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	cr := CommentApiResponse{}

	if err = json.Unmarshal(buf, &cr); err != nil {
		log.Fatal(err)
	}

	return cr
}

func GetComments(articleID string) []Comment {
	csrs := commentsRepliesNested(getAllComments(articleID))
	return csrs.Flatten()
}
