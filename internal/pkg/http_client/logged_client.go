package http_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LoggedClient struct {
	*http.Client
	logServerURL string
}

type LogEntry struct {
	Timestamp    string            `json:"timestamp"`
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers"`
	RequestBody  string            `json:"request_body"`
	StatusCode   int               `json:"status_code"`
	ResponseBody string            `json:"response_body"`
	Duration     int64             `json:"duration_ms"`
	Error        string            `json:"error,omitempty"`
}

func NewLoggedClient(logServerURL string) *LoggedClient {
	return &LoggedClient{
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logServerURL: logServerURL,
	}
}

func (c *LoggedClient) Do(req *http.Request) (*http.Response, error) {
	startTime := time.Now()

	// Копируем тело запроса для логирования
	var requestBody []byte
	if req.Body != nil {
		requestBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}

	// Копируем заголовки (скрываем токены)
	headers := make(map[string]string)
	for key, values := range req.Header {
		if key == "Authorization" {
			headers[key] = "[REDACTED]"
		} else if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Выполняем запрос
	resp, err := c.Client.Do(req)

	// Подготавливаем лог
	logEntry := LogEntry{
		Timestamp:   time.Now().Format(time.RFC3339),
		Method:      req.Method,
		URL:         req.URL.String(),
		Headers:     headers,
		RequestBody: string(requestBody),
		Duration:    time.Since(startTime).Milliseconds(),
	}

	if err != nil {
		logEntry.Error = err.Error()
		c.sendLog(logEntry)
		return nil, err
	}

	// Копируем тело ответа
	responseBody, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	logEntry.StatusCode = resp.StatusCode
	logEntry.ResponseBody = string(responseBody)

	// Отправляем лог асинхронно
	go c.sendLog(logEntry)

	return resp, nil
}

func (c *LoggedClient) sendLog(entry LogEntry) {
	if c.logServerURL == "" {
		// Если нет сервера логов, печатаем в консоль
		fmt.Printf("[HTTP] %s %s -> %d (%dms)\n",
			entry.Method, entry.URL, entry.StatusCode, entry.Duration)
		return
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return
	}

	// Отправляем лог на сервер
	http.Post(c.logServerURL+"/log", "application/json", bytes.NewBuffer(jsonData))
}
