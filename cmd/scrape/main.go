package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pbaettig/20miner/internal/config"
	"github.com/pbaettig/20miner/internal/pkg/articles"
)

func main() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	articleLinks, err := articles.GetArticleLinks(client, config.FrontURL)
	if err != nil {
		log.Fatal(err)
	}

	// for _, al := range articleLinks {
	// 	fmt.Printf("%+v\n", al)
	// }

	fmt.Println()
	fmt.Printf("%+v\n", articleLinks[1])
	fmt.Printf("%+v\n", articleLinks[1].Get(client).Title)

}
