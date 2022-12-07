package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	URL_API				 = "https://www.tabnews.com.br/api/v1"
	URL_CONTENTS   = URL_API + "/contents"
	PAGE_SIZE			 = 40
)

var (
	currentPage = 1
	currentStrategy = "relevant"
)

func DownloadContent() ([]Content, error) {
	// Return cached results if exist
	if len(contents) > 0 && len(cachedContents) > 0 {
		content := cachedContents[currentPage]
		if len(content) > 0 {
			return content, nil
		}
	}

	// Perform HTTP request to load results
	resp, err := http.Get(fmt.Sprintf("%s%s%d%s%d%s%s", URL_CONTENTS, "?per_page=", PAGE_SIZE, "&page=", currentPage, "&strategy=",  currentStrategy))

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}

	defer resp.Body.Close()

	var content = []Content{}
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	decoder.Decode(&content)

	// Save page results into cache
	cachedContents[currentPage] = content

	return content, nil
}

func DownloadArticle(username string, slug string, id string) (*Article, error) {
	// Return cached result if exist
	if len(cachedArticles) > 0 {
		article := cachedArticles[id]
		if article != nil {
			return article, nil
		}
	}

	// Perform HTTP request to load results
	resp, err := http.Get(URL_CONTENTS + "/" + username + "/" + slug)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}

	defer resp.Body.Close()

	var article = Article{}
	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()
	decoder.Decode(&article)

	// Save article into cache
	cachedArticles[id] = &article

	return &article, nil
}