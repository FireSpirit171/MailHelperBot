package domain

import "time"

type GroupSession struct {
	GroupID          int64     `json:"group_id"`
	GroupTitle       string    `json:"group_title"`
	OwnerID          int64     `json:"owner_id"`
	MediaType        string    `json:"media_type"` // "photos", "videos", "all"
	HistoryProcessed bool      `json:"history_processed"`
	CreatedAt        time.Time `json:"created_at"`
}

type ProcessedMedia struct {
	MediaID   string    `json:"media_id"`
	GroupID   int64     `json:"group_id"`
	FileID    string    `json:"file_id"`
	FileName  string    `json:"file_name"`
	MediaType string    `json:"media_type"`
	CreatedAt time.Time `json:"created_at"`
}
