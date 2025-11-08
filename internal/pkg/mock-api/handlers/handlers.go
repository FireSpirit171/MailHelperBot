package handlers

import (
	"encoding/json"
	"fmt"
	"mail_helper_bot/internal/pkg/mock-api/models"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// MkdirHandler обрабатывает создание папки
func MkdirHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем путь из URL
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/private/mkdir/")
	if path == "" {
		sendError(w, "Path is required", http.StatusBadRequest)
		return
	}

	// Создаем mock response
	response := models.MkdirResponse{
		Hidden: false,
		Kind:   "folder",
		Mtime:  int(time.Now().Unix()),
		Name:   path,
		NodeID: generateID(),
		Path:   "/" + path,
		Size:   0,
		Type:   "folder",
		Views:  0,
	}

	response.Attributes.Actor = "mock_user"
	response.Attributes.Grantor = "system"
	response.Attributes.Mandate = "full_access"

	response.Counts.Files = 0
	response.Counts.Folders = 0

	response.DownloadLimit.Left = 100
	response.DownloadLimit.NextReset = int(time.Now().Add(24 * time.Hour).Unix())
	response.DownloadLimit.Total = 100

	response.Flags.Blocked = false
	response.Flags.Depo = true
	response.Flags.Favorite = false
	response.Flags.Restricted = false

	response.Link = generateLink(path)
	response.List = []string{}

	response.Malware.Status = "clean"

	response.Thumb.Xm0 = ""
	response.Thumb.Xms0 = ""
	response.Thumb.Xms4 = ""

	sendJSON(w, response)
}

// AddHandler обрабатывает добавление файла
func AddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request models.AddRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация
	if request.Hash == "" || request.Path == "" {
		sendError(w, "Hash and path are required", http.StatusBadRequest)
		return
	}

	// Формируем ответ с теми же данными
	response := models.AddRequest{
		Hash:            request.Hash,
		LastMod:         request.LastMod,
		Overwrite:       request.Overwrite,
		Path:            request.Path,
		Size:            request.Size,
		UnlimitedUpload: true,
		UploadType:      "standard",
	}

	response.Options.AutoRename = true
	response.Options.ExclByHash = true
	response.Options.Overwrite = request.Overwrite

	sendJSON(w, response)
}

// ShareHandler обрабатывает создание публичной ссылки
func ShareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем путь из URL
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/private/share/")
	if path == "" {
		sendError(w, "Path is required", http.StatusBadRequest)
		return
	}

	response := generateLink(path)
	response.URL = fmt.Sprintf("https://mock-storage.example.com/share/%s", response.ID)

	sendJSON(w, response)
}

// UnshareHandler обрабатывает удаление публичной ссылки
func UnshareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем путь из URL
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/private/unshare/")
	if path == "" {
		sendError(w, "Path is required", http.StatusBadRequest)
		return
	}

	response := generateLink(path)
	response.URL = "" // URL пустой при unshare

	sendJSON(w, response)
}

// Вспомогательные функции
func generateID() string {
	return fmt.Sprintf("%d%d", time.Now().Unix(), rand.Intn(10000))
}

func generateLink(name string) models.Link {
	link := models.Link{
		Ctime:     int(time.Now().Unix()),
		Downloads: 0,
		Expires:   int(time.Now().Add(7 * 24 * time.Hour).Unix()),
		ExtID:     generateID(),
		ID:        generateID(),
		Mode:      "read",
		Name:      name,
		Owner:     true,
		Type:      "folder",
		Unknown:   false,
		URL:       "",
		Views:     0,
	}

	link.Flags.SEOIndexed = false
	link.Flags.Commentable = true
	link.Flags.Domestic = true
	link.Flags.EmailListAccess = false
	link.Flags.Writable = false

	return link
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errResp := models.ErrorResponse{
		Error:  message,
		Fields: []string{},
	}

	json.NewEncoder(w).Encode(errResp)
}
