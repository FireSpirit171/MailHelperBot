package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type LogStorage struct {
	mu   sync.Mutex
	logs []map[string]interface{}
	file *os.File
}

var storage *LogStorage

func main() {
	// Создаем папку для логов
	os.MkdirAll("/logs", 0755)

	// Открываем файл для записи
	logFile, err := os.OpenFile(
		fmt.Sprintf("/logs/http_%s.log", time.Now().Format("2006-01-02")),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	storage = &LogStorage{
		logs: make([]map[string]interface{}, 0),
		file: logFile,
	}

	// Endpoints
	http.HandleFunc("/log", handleLog)
	http.HandleFunc("/logs", handleGetLogs)
	http.HandleFunc("/health", handleHealth)

	port := "8081"
	log.Printf("Log server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal(body, &logEntry); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	storage.mu.Lock()
	defer storage.mu.Unlock()

	// Сохраняем в памяти (последние 1000 записей)
	storage.logs = append(storage.logs, logEntry)
	if len(storage.logs) > 1000 {
		storage.logs = storage.logs[1:]
	}

	// Записываем в файл
	storage.file.WriteString(string(body) + "\n")
	storage.file.Sync()

	// Печатаем в консоль для отладки
	method := logEntry["method"]
	url := logEntry["url"]
	status := logEntry["status_code"]
	duration := logEntry["duration_ms"]

	log.Printf("[%s] %s -> %v (%vms)", method, url, status, duration)

	if logEntry["error"] != nil {
		log.Printf("  ERROR: %s", logEntry["error"])
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleGetLogs(w http.ResponseWriter, r *http.Request) {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(storage.logs)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
