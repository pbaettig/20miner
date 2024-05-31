package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type ArticleLink struct {
	ID    string
	Href  string
	Title string
}

func (al ArticleLink) Get(c *http.Client) Article {
	a := Article{ArticleLink: al}

	uri, err := url.JoinPath("https://www.20min.ch", al.Href)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(uri)

	resp, err := c.Get(uri)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	recurseNodes(doc, func(n *html.Node) {
		if n.DataAtom == atom.Time && n.PrevSibling.DataAtom == atom.Span {
			fmt.Printf("%+v\n", n.FirstChild)
			pt, _ := time.Parse(time.RFC3339, getHtmlNodeAttr(n, "datetime"))
			a.PublicationDate = pt
			// a.PublicationDate = n.FirstChild.Data
		}

	})

	return a
}

type Article struct {
	ArticleLink

	PublicationDate time.Time
	Category        string
}

func recurseNodes(n *html.Node, callbackFn func(*html.Node)) {
	if n.Type == html.ElementNode {
		callbackFn(n)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		recurseNodes(c, callbackFn)
	}
}

func getHtmlNodeAttr(n *html.Node, k string) string {
	for _, att := range n.Attr {
		if att.Key == k {
			return att.Val
		}
	}

	return ""
}

func processArticleLinkNode(n *html.Node) (a ArticleLink) {
	a.Href = getHtmlNodeAttr(n, "href")
	a.ID = getHtmlNodeAttr(n.Parent, "id")

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

func processHTMLNode(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "a" && n.Parent.Data == "article" {
		fmt.Println(processArticleLinkNode(n))
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		processHTMLNode(c)
	}

}

func main() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("https://www.20min.ch/front")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	articles := make([]ArticleLink, 0)
	recurseNodes(doc, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" && n.Parent.Data == "article" {
			a := processArticleLinkNode(n)
			// fmt.Println(a)
			articles = append(articles, a)
		}
	})

	for i, a := range articles {
		fmt.Println(i, a)
	}
	fmt.Printf("%+v\n", articles[0].Get(client))

}
