package web_server_service

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"mail_helper_bot/internal/pkg/oauth/oauth_service"
	"net/http"
)

// todo: прописать нормальные интерфейсы
type WebServer struct {
	oauth  *oauth_service.OAuthService
	botAPI *tgbotapi.BotAPI
	port   string
}

func NewWebServer(oauth *oauth_service.OAuthService, botAPI *tgbotapi.BotAPI, port string) *WebServer {
	return &WebServer{
		oauth:  oauth,
		botAPI: botAPI,
		port:   port,
	}
}

// todo: вынести в отдельный main и ручки в router
func (ws *WebServer) Start() error {
	http.HandleFunc("/oauth/callback/", ws.handleOAuthCallback)
	http.HandleFunc("/health/", ws.handleHealthCheck)

	log.Printf("Starting web server on port %s", ws.port)
	return http.ListenAndServe(":"+ws.port, nil)
}

func (ws *WebServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (ws *WebServer) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	log.Printf("OAuth callback received: %s", r.URL.String())

	queryParams := r.URL.Query()
	code := queryParams.Get("code")
	state := queryParams.Get("state")
	errorParam := queryParams.Get("error")

	if errorParam != "" {
		errorDescription := queryParams.Get("error_description")
		log.Printf("OAuth error: %s - %s", errorParam, errorDescription)
		ws.sendErrorResponse(w, "Ошибка авторизации: "+errorDescription)
		return
	}

	if code == "" {
		ws.sendErrorResponse(w, "Код авторизации не получен")
		return
	}

	if state == "" {
		ws.sendErrorResponse(w, "State параметр отсутствует")
		return
	}

	chatID, err := ws.oauth.ValidateState(state)
	log.Printf("chatID seccessfully extracted - %v", chatID)
	if err != nil {
		log.Printf("Invalid state: %v", err)
		ws.sendErrorResponse(w, "Неверный или просроченный state параметр")
		return
	}

	tokenResp, err := ws.oauth.ExchangeCodeForToken(code, state)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		ws.sendErrorResponse(w, "Ошибка при получении токена")
		return
	}

	userInfo, err := ws.oauth.GetUserInfo(tokenResp.AccessToken)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		ws.sendErrorResponse(w, "Ошибка при получении информации о пользователе")
		return
	}

	err = ws.oauth.SaveUserSession(chatID, tokenResp, userInfo)
	if err != nil {
		log.Printf("Error saving session: %v", err)
		ws.sendErrorResponse(w, "Ошибка при сохранении сессии")
		return
	}

	successMsg := fmt.Sprintf("✅ Авторизация успешна!\n\n👤 Имя: %s\n📧 Email: %s",
		userInfo.Name, userInfo.Email)

	msg := tgbotapi.NewMessage(chatID, successMsg)
	_, err = ws.botAPI.Send(msg)
	if err != nil {
		log.Printf("Error sending Telegram message: %v", err)
	}

	ws.sendGroupAdditionInstructions(chatID, userInfo.Name)

	ws.sendSuccessResponse(w, userInfo.Name, userInfo.Email)
}

func (ws *WebServer) sendGroupAdditionInstructions(chatID int64, userName string) {
	// Даем небольшую задержку, чтобы сообщение об успешной авторизации успело отправиться
	// time.Sleep(2 * time.Second) // Можно раскомментировать если нужно

	instructions := fmt.Sprintf(`🎉 **Отлично, %s! Теперь давайте добавим бота в группу!**

👥 **Как добавить бота в группу:**

1. **Откройте любую вашу группу** в Telegram
2. **Нажмите на название группы** вверху экрана
3. **Выберите "Участники"**
4. **Нажмите "Добавить участников"**
5. **Найдите @%s** и добавьте бота
6. **Сделайте бота администратором:**
   - ✅ Чтение сообщений
   - ✅ Доступ к медиафайлам
   - ✅ Отправка сообщений

🎯 **После добавления бота в группу:**
• Бот автоматически определит новую группу
• Отправит сообщение с настройками
• Предложит выбрать тип медиа для загрузки (фото/видео/все)

💡 **Советы:**
• Можно добавить в несколько групп
• Каждая группа настраивается отдельно
• Используйте /my_groups для просмотра всех ваших групп

🚀 **Добавьте бота в группу и начинайте выгрузку медиа в ваше облако!**`, userName, ws.botAPI.Self.UserName)

	msg := tgbotapi.NewMessage(chatID, instructions)
	msg.ParseMode = "Markdown"

	if _, err := ws.botAPI.Send(msg); err != nil {
		log.Printf("Failed to send group addition instructions to %d: %v", chatID, err)
	}
}

// todo: это пздец
func (ws *WebServer) sendSuccessResponse(w http.ResponseWriter, name, email string) {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Успешная авторизация</title>
    <meta charset="utf-8">
    <style>
        body { 
            font-family: Arial, sans-serif; 
            text-align: center; 
            padding: 50px; 
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
        }
        .container { 
            background: rgba(255,255,255,0.1); 
            padding: 30px; 
            border-radius: 15px;
            backdrop-filter: blur(10px);
            max-width: 500px;
            margin: 0 auto;
        }
        .success { 
            font-size: 24px; 
            margin-bottom: 20px;
        }
        .info {
            background: rgba(255,255,255,0.2);
            padding: 15px;
            border-radius: 8px;
            margin: 15px 0;
        }
        .close-btn {
            background: #4CAF50;
            color: white;
            border: none;
            padding: 12px 30px;
            border-radius: 25px;
            cursor: pointer;
            font-size: 16px;
            margin-top: 20px;
        }
        .close-btn:hover {
            background: #45a049;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success">✅ Авторизация успешна!</div>
        <div class="info">
            <strong>👤 Имя:</strong> %s<br>
            <strong>📧 Email:</strong> %s
        </div>
        <p>Вы можете закрыть эту страницу и вернуться в Telegram</p>
        <button class="close-btn" onclick="window.close()">Закрыть окно</button>
    </div>
</body>
</html>`, name, email)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (ws *WebServer) sendErrorResponse(w http.ResponseWriter, errorMsg string) {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Ошибка авторизации</title>
    <meta charset="utf-8">
    <style>
        body { 
            font-family: Arial, sans-serif; 
            text-align: center; 
            padding: 50px; 
            background: linear-gradient(135deg, #ff6b6b 0%%, #ee5a24 100%%);
            color: white;
        }
        .container { 
            background: rgba(255,255,255,0.1); 
            padding: 30px; 
            border-radius: 15px;
            backdrop-filter: blur(10px);
            max-width: 500px;
            margin: 0 auto;
        }
        .error { 
            font-size: 24px; 
            margin-bottom: 20px;
        }
        .message {
            background: rgba(255,255,255,0.2);
            padding: 15px;
            border-radius: 8px;
            margin: 15px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="error">❌ Ошибка авторизации</div>
        <div class="message">%s</div>
        <p>Пожалуйста, попробуйте еще раз</p>
    </div>
</body>
</html>`, errorMsg)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}
