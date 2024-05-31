package articles

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pbaettig/20miner/internal/config"
	hp "github.com/pbaettig/20miner/internal/pkg/htmlparser"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type ArticleLink struct {
	ID    string
	Href  string
	Title string
}

type Article struct {
	ArticleLink

	PublicationDate time.Time
	Category        string
	Text            string
}

func (al ArticleLink) Get(c *http.Client) Article {
	a := Article{ArticleLink: al}

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

		if n.DataAtom == atom.Span && n.DataAtom == atom.Button {
			class := hp.GetNodeAttr(n, "class")
			if strings.HasPrefix(class, "ActivityButton_activityCounter_") {
				fmt.Printf("************** Class: %s\n", class)
				fmt.Printf("************** Parent: %+v\n", n.Parent)
				fmt.Printf("************** Acitivity: %s\n", hp.GetNodeChildData(n))
				fmt.Println()
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
	a.ID = hp.GetNodeAttr(n.Parent, "id")

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