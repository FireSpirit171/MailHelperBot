package domain

import "time"

type GroupSession struct {
	GroupID          int64     `json:"group_id"`
	GroupTitle       string    `json:"group_title"`
	OwnerChatID      int64     `json:"owner_id"`
	MediaType        string    `json:"media_type"` // "photos", "videos", "all"
	CloudFolderPath  string    `json:"cloud_folder_path"`
	PublicURL        string    `json:"public_url"`
	HistoryProcessed bool      `json:"history_processed"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ProcessedMedia struct {
	ID            string    `json:"id"`
	GroupID       int64     `json:"group_id"`
	FileUniqueID  string    `json:"file_unique_id"`
	FileName      string    `json:"file_name"`
	MediaType     string    `json:"media_type"`
	FileSizeBytes int64     `json:"file_size_bytes"`
	UploadedAt    time.Time `json:"uploaded_at"`
}

type GroupStats struct {
	PhotosCount    int
	VideosCount    int
	TotalSizeBytes int64
	MediaType      string
	PublicURL      string
}

type SharedFolder struct {
	ID         int
	ChatID     int64
	FolderName string
	FolderPath string
	PublicURL  string
	CreatedAt  time.Time
}
