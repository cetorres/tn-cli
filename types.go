package main

import (
	"time"
)

type Content struct {
	ID                string      `json:"id"`
	OwnerID           string      `json:"owner_id"`
	ParentID          interface{} `json:"parent_id"`
	Slug              string      `json:"slug"`
	Title             string      `json:"title"`
	Status            string      `json:"status"`
	SourceURL         interface{} `json:"source_url"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	PublishedAt       time.Time   `json:"published_at"`
	DeletedAt         interface{} `json:"deleted_at"`
	Tabcoins          int         `json:"tabcoins"`
	OwnerUsername     string      `json:"owner_username"`
	ChildrenDeepCount int         `json:"children_deep_count"`
}

type Article struct {
	ID                string      `json:"id"`
	OwnerID           string      `json:"owner_id"`
	ParentID          interface{} `json:"parent_id"`
	Slug              string      `json:"slug"`
	Title             string      `json:"title"`
	Body              string      `json:"body"`
	Status            string      `json:"status"`
	SourceURL         interface{} `json:"source_url"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	PublishedAt       time.Time   `json:"published_at"`
	DeletedAt         interface{} `json:"deleted_at"`
	OwnerUsername     string      `json:"owner_username"`
	Tabcoins          int         `json:"tabcoins"`
	ChildrenDeepCount int         `json:"children_deep_count"`
}