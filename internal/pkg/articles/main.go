package articles

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pbaettig/20miner/internal/config"
	"github.com/pbaettig/20miner/internal/pkg/comments"
	hp "github.com/pbaettig/20miner/internal/pkg/htmlparser"
	"github.com/pbaettig/20miner/internal/pkg/utils"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"gorm.io/gorm"
)

type ArticleLink struct {
	OriginalID string
	Href       string
	Title      string
}

type Shares struct {
	gorm.Model

	ArticleID uint
	Value     uint
}

type Article struct {
	ArticleLink
	gorm.Model
	PublicationDate time.Time
	Category        string
	Text            string
	Shares          Shares
	Comments        []*comments.Comment
}

// Get full article details from Article Link by downloading the page
// and parsing the HTML
func (al ArticleLink) Get(c *http.Client) Article {
	a := Article{ArticleLink: al}
	a.ID = utils.MustIntToUint(al.OriginalID)

	uri, err := url.JoinPath(config.BaseURL, al.Href)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := hp.RootNodeFromURL(c, uri)
	if err != nil {
		log.Fatal(err)
	}

	hp.VisitNodes(doc, func(n *html.Node, siblings []*html.Node) {
		if n.DataAtom == atom.Time && n.PrevSibling.DataAtom == atom.Span {
			pt, _ := time.Parse(time.RFC3339, hp.GetNodeAttr(n, "datetime"))
			a.PublicationDate = pt
		}

		if n.DataAtom == atom.A && n.Parent.DataAtom == atom.Div && n.Parent.Parent.DataAtom == atom.Section {
			href := hp.GetNodeAttr(n, "href")
			if href != "" && href != "/" {
				a.Category = n.FirstChild.Data
			}
		}

		if n.DataAtom == atom.Span && n.Parent.DataAtom == atom.Button {
			parentTestID := hp.GetNodeAttr(n.Parent, "data-testid")
			if parentTestID == "ButtonShare" && n.FirstChild.Type == html.TextNode {
				a.Shares = Shares{Value: utils.MustIntToUint(hp.GetNodeChildData(n))}
			}
		}

		if n.DataAtom == atom.Section && n.Parent.DataAtom == atom.Article {
			articleSections := make([]string, 0)

			hp.VisitNodes(n, func(n *html.Node, siblings []*html.Node) {
				if n.DataAtom == atom.P {
					if t := hp.GetNodeChildData(n); t != "" {
						articleSections = append(articleSections, t)
					}

				}
			})

			a.Text = strings.Join(articleSections, "\n")
		}
	})

	return a
}

func processArticleLinkNode(n *html.Node) (a ArticleLink) {
	a.Href = hp.GetNodeAttr(n, "href")
	a.OriginalID = hp.GetNodeAttr(n.Parent, "id")

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "span" && n.Parent.Data == "h2" {
			a.Title = n.FirstChild.Data
			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return
}

// GetArticleLinks prepares a list of all articles found on the frontpage
func GetArticleLinks(c *http.Client, uri string) (articles []ArticleLink, err error) {
	doc, err := hp.RootNodeFromURL(c, uri)
	if err != nil {
		return
	}

	hp.VisitNodes(doc, func(n *html.Node, siblings []*html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" && n.Parent.Data == "article" {
			articles = append(articles, processArticleLinkNode(n))
		}
	})

	return
}
