package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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
	log.Printf("📊 Log server starting on port %s", port)
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

	// Красивый вывод в консоль
	printLogEntry(logEntry)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func printLogEntry(entry map[string]interface{}) {
	// Цвета для терминала
	const (
		colorReset  = "\033[0m"
		colorRed    = "\033[31m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorPurple = "\033[35m"
		colorCyan   = "\033[36m"
		colorGray   = "\033[37m"
		colorBold   = "\033[1m"
	)

	id := entry["id"]
	method := entry["method"]
	url := entry["url"]
	statusCode := entry["status_code"]
	duration := entry["duration_ms"]
	timestamp := entry["timestamp"]

	// Определяем цвет для статуса
	statusColor := colorGray
	if statusCode != nil {
		status := int(statusCode.(float64))
		if status >= 200 && status < 300 {
			statusColor = colorGreen
		} else if status >= 400 && status < 500 {
			statusColor = colorYellow
		} else if status >= 500 {
			statusColor = colorRed
		}
	}

	// Заголовок запроса
	fmt.Printf("\n%s========== HTTP REQUEST [%s] ==========%s\n",
		colorBold+colorCyan, id, colorReset)
	fmt.Printf("%sTime:%s %s\n", colorBold, colorReset, timestamp)
	fmt.Printf("%sMethod:%s %s%s%s\n", colorBold, colorReset, colorBlue, method, colorReset)
	fmt.Printf("%sURL:%s %s\n", colorBold, colorReset, url)

	// Заголовки
	if headers, ok := entry["headers"].(map[string]interface{}); ok && len(headers) > 0 {
		fmt.Printf("%sHeaders:%s\n", colorBold, colorReset)
		for key, value := range headers {
			fmt.Printf("  %s%s:%s %v\n", colorGray, key, colorReset, value)
		}
	}

	// Тело запроса
	if reqBody, ok := entry["request_body"].(string); ok && reqBody != "" {
		fmt.Printf("%sRequest Body:%s\n", colorBold, colorReset)
		printFormattedBody(reqBody, "  ")
	}

	// Статус ответа
	if statusCode != nil {
		fmt.Printf("%sStatus Code:%s %s%v%s\n",
			colorBold, colorReset, statusColor, statusCode, colorReset)
	}

	// Тело ответа
	if respBody, ok := entry["response_body"].(string); ok && respBody != "" {
		fmt.Printf("%sResponse Body:%s\n", colorBold, colorReset)
		printFormattedBody(respBody, "  ")
	}

	// Ошибка
	if errorMsg, ok := entry["error"].(string); ok && errorMsg != "" {
		fmt.Printf("%s%sError:%s %s\n", colorRed, colorBold, colorReset, errorMsg)
	}

	// Длительность
	if duration != nil {
		durationColor := colorGreen
		dur := int64(duration.(float64))
		if dur > 1000 {
			durationColor = colorYellow
		}
		if dur > 5000 {
			durationColor = colorRed
		}
		fmt.Printf("%sDuration:%s %s%dms%s\n",
			colorBold, colorReset, durationColor, dur, colorReset)
	}

	fmt.Printf("%s========================================%s\n",
		colorCyan, colorReset)
}

func printFormattedBody(body string, indent string) {
	// Пытаемся распарсить как JSON для красивого вывода
	var jsonData interface{}
	if err := json.Unmarshal([]byte(body), &jsonData); err == nil {
		prettyJSON, _ := json.MarshalIndent(jsonData, indent, "  ")
		fmt.Printf("%s%s\n", indent, string(prettyJSON))
	} else {
		// Если не JSON, выводим как есть
		lines := strings.Split(body, "\n")
		for _, line := range lines {
			if len(line) > 200 {
				fmt.Printf("%s%s... [truncated]\n", indent, line[:200])
			} else {
				fmt.Printf("%s%s\n", indent, line)
			}
		}
	}
}

func handleGetLogs(w http.ResponseWriter, r *http.Request) {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	// Фильтрация по параметрам
	urlFilter := r.URL.Query().Get("url")
	methodFilter := r.URL.Query().Get("method")
	statusFilter := r.URL.Query().Get("status")

	filteredLogs := []map[string]interface{}{}
	for _, log := range storage.logs {
		// Применяем фильтры
		if urlFilter != "" {
			if url, ok := log["url"].(string); !ok || !strings.Contains(url, urlFilter) {
				continue
			}
		}
		if methodFilter != "" {
			if method, ok := log["method"].(string); !ok || method != methodFilter {
				continue
			}
		}
		if statusFilter != "" {
			if status, ok := log["status_code"].(float64); !ok || fmt.Sprintf("%d", int(status)) != statusFilter {
				continue
			}
		}

		filteredLogs = append(filteredLogs, log)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filteredLogs)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
