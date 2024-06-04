package htmlparser

import (
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

func RootNodeFromURL(c *http.Client, url string) (n *html.Node, err error) {
	resp, err := c.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	n, err = html.Parse(resp.Body)

	return
}

func VisitNodes(n *html.Node, callbackFn func(n *html.Node, siblings []*html.Node)) {
	if n == nil {
		return
	}

	siblings := make([]*html.Node, 0)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		siblings = append(siblings, c)
	}

	if n.Type == html.ElementNode {
		callbackFn(n, siblings)
	}

	for _, s := range siblings {
		VisitNodes(s, callbackFn)
	}
}

func GetNodeChildData(n *html.Node) (d string) {
	if n == nil {
		return
	}

	if n.FirstChild == nil {
		return
	}

	d = strings.TrimSpace(n.FirstChild.Data)
	return
}

func GetNodeAttr(n *html.Node, k string) string {
	if n == nil {
		return ""
	}

	for _, att := range n.Attr {
		if att.Key == k {
			return strings.TrimSpace(att.Val)
		}
	}

	return ""
}
