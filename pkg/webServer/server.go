package webServer

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"reminder/configs"
	"reminder/internal/messages"
	"reminder/internal/tasks"
	"strconv"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// WebServer представляет веб-сервер для работы с веб-приложением выбора даты
type WebServer struct {
	tasksRepo *tasks.Repository
	msgRepo   *messages.Repository
	bot       *gotgbot.Bot
	config    *configs.AppConfig
	server    *http.Server
}

// NewWebServer создает новый экземпляр WebServer
func NewWebServer(tr *tasks.Repository, ms *messages.Repository, bot *gotgbot.Bot, cfg *configs.AppConfig) *WebServer {
	return &WebServer{
		tasksRepo: tr,
		msgRepo:   ms,
		bot:       bot,
		config:    cfg,
	}
}

// Run Конфигурирует и запускает сервер
func (ws *WebServer) Run(ctx context.Context) error {
	r := http.NewServeMux()
	fs := http.FileServer(http.Dir("datePicker/assets"))
	r.Handle("GET /assets/", http.StripPrefix("/assets/", fs))
	r.HandleFunc("GET /app", ws.webAppHandler)
	r.HandleFunc("POST /date", ws.setDate)

	// Создаем HTTP сервер
	addr := ":" + ws.config.ServerPort
	ws.server = &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		var err error
		if ws.config.UseTLS {
			log.Printf("Запускаем HTTPS сервер на %s", addr)
			err = ws.server.ListenAndServeTLS(ws.config.TLSCertPath, ws.config.TLSKeyPath)
		} else {
			log.Printf("Запускаем HTTP сервер на %s", addr)
			err = ws.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Printf("Ошибка запуска веб-сервера: %v", err)
		}
	}()

	// Ожидаем сигнал остановки
	<-ctx.Done()

	// Graceful shutdown
	log.Println("Останавливаем веб-сервер...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ws.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Ошибка при остановке сервера: %v", err)
		return err
	}

	log.Println("Веб-сервер остановлен")
	return nil
}

// webAppHandler обработчик, отвечающий за приложение селектора даты напоминания
func (ws *WebServer) webAppHandler(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join("datePicker", "index.html")
	http.ServeFile(w, r, path)
}

// setDate обработчик запроса установки даты напоминания. Запрос приходит из web app.
func (ws *WebServer) setDate(w http.ResponseWriter, r *http.Request) {
	// парсим пришедшие данные
	payload := struct {
		Date   string `json:"date"`
		TaskId string `json:"taskId"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("Ошибка декодирования JSON: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	taskId, err := strconv.Atoi(payload.TaskId)
	if err != nil {
		log.Printf("Ошибка парсинга TaskId: %v", err)
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	newDate, err := time.Parse(time.RFC3339, payload.Date)
	if err != nil {
		log.Printf("Ошибка парсинга даты: %v", err)
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}
	newDate = newDate.UTC()

	// если данные успешно распаршены, добавляем дату напоминанию
	err = ws.tasksRepo.SetDate(taskId, &newDate)
	if err != nil {
		log.Printf("Ошибка сохранения даты: %v", err)
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	// затем, нужно пройтись по всем сообщениям чата, имеющим предложение установить дату,
	// и у этих сообщений удаляем кнопку установки даты.
	// также, нужно удалить из бд все записи-связи напоминания и сообщений с предложением установить дату
	go func() {
		messageList, err := ws.msgRepo.GetAllByTaskID(taskId)
		if err != nil {
			log.Printf("Ошибка получения сообщений: %v", err)
			return
		}

		for _, m := range messageList {
			_, _, err = ws.bot.EditMessageReplyMarkup(&gotgbot.EditMessageReplyMarkupOpts{
				MessageId:   int64(m.MessageID),
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: nil},
				ChatId:      int64(m.ChatID),
			})
			if err != nil {
				log.Printf("Ошибка редактирования разметки сообщения: %v", err)
			}
		}

		err = ws.msgRepo.DeleteAllByTaskId(taskId)
		if err != nil {
			log.Printf("Ошибка удаления сообщений: %v", err)
		}
	}()

	w.WriteHeader(http.StatusOK)
}
