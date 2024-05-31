package config

import "fmt"

const (
	BaseURL = "https://www.20min.ch"
)

var (
	FrontURL       = fmt.Sprintf("%s/front", BaseURL)
	CommentsAPIURL = "https://api.20min.ch/comment/v1/comments"
)
