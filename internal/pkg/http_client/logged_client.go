package http_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LoggedClient struct {
	*http.Client
	logServerURL string
}

type LogEntry struct {
	ID           string              `json:"id"`
	Timestamp    string              `json:"timestamp"`
	Method       string              `json:"method"`
	URL          string              `json:"url"`
	Headers      map[string][]string `json:"headers"`
	RequestBody  string              `json:"request_body"`
	StatusCode   int                 `json:"status_code"`
	ResponseBody string              `json:"response_body"`
	Duration     int64               `json:"duration_ms"`
	Error        string              `json:"error,omitempty"`
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
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Копируем тело запроса для логирования
	var requestBody []byte
	if req.Body != nil {
		requestBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}

	// Копируем ВСЕ заголовки
	headers := make(map[string][]string)
	for key, values := range req.Header {
		headers[key] = values
	}

	// Выполняем запрос
	resp, err := c.Client.Do(req)

	// Подготавливаем лог
	logEntry := LogEntry{
		ID:          requestID,
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
		c.printToConsole(entry)
		return
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return
	}

	// Отправляем лог на сервер
	http.Post(c.logServerURL+"/log", "application/json", bytes.NewBuffer(jsonData))
}

func (c *LoggedClient) printToConsole(entry LogEntry) {
	fmt.Printf("\n========== HTTP REQUEST [%s] ==========\n", entry.ID)
	fmt.Printf("Time: %s\n", entry.Timestamp)
	fmt.Printf("Method: %s\n", entry.Method)
	fmt.Printf("URL: %s\n", entry.URL)

	if len(entry.Headers) > 0 {
		fmt.Println("Headers:")
		for key, values := range entry.Headers {
			fmt.Printf("  %s: %s\n", key, strings.Join(values, ", "))
		}
	}

	if entry.RequestBody != "" {
		fmt.Println("Request Body:")
		// Пытаемся красиво отформатировать JSON
		var jsonData interface{}
		if err := json.Unmarshal([]byte(entry.RequestBody), &jsonData); err == nil {
			prettyJSON, _ := json.MarshalIndent(jsonData, "  ", "  ")
			fmt.Printf("  %s\n", string(prettyJSON))
		} else {
			fmt.Printf("  %s\n", entry.RequestBody)
		}
	}

	if entry.StatusCode > 0 {
		fmt.Printf("Status Code: %d\n", entry.StatusCode)
	}

	if entry.ResponseBody != "" {
		fmt.Println("Response Body:")
		// Пытаемся красиво отформатировать JSON
		var jsonData interface{}
		if err := json.Unmarshal([]byte(entry.ResponseBody), &jsonData); err == nil {
			prettyJSON, _ := json.MarshalIndent(jsonData, "  ", "  ")
			fmt.Printf("  %s\n", string(prettyJSON))
		} else {
			// Если не JSON, выводим как есть (обрезаем если слишком длинный)
			if len(entry.ResponseBody) > 1000 {
				fmt.Printf("  %s... [truncated]\n", entry.ResponseBody[:1000])
			} else {
				fmt.Printf("  %s\n", entry.ResponseBody)
			}
		}
	}

	if entry.Error != "" {
		fmt.Printf("Error: %s\n", entry.Error)
	}

	fmt.Printf("Duration: %dms\n", entry.Duration)
	fmt.Println("=======================================\n")
}
