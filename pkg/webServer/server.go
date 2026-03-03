package webServer

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"reminder/internal/messages"
	"reminder/internal/tasks"
	"runtime"
	"strconv"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type WebServer struct {
	tasksRepo *tasks.Repository
	msgRepo   *messages.Repository
	bot       *gotgbot.Bot
}

func NewWebServer(tr *tasks.Repository, ms *messages.Repository, bot *gotgbot.Bot) *WebServer {
	return &WebServer{tasksRepo: tr, msgRepo: ms, bot: bot}
}

// Run Конфигурирует и запускает сервак
func (ws *WebServer) Run() {
	r := http.NewServeMux()
	fs := http.FileServer(http.Dir("datePicker/assets"))
	r.Handle("GET /assets/", http.StripPrefix("/assets/", fs))
	r.HandleFunc("GET /app", ws.webAppHandler)
	r.HandleFunc("POST /date", ws.setDate)

	var err error

	// условие для запуска сервера локально и на сервере
	if runtime.GOOS == "windows" {
		err = http.ListenAndServe(":3000", r)
	} else {
		err = http.ListenAndServeTLS(
			":443",
			"/etc/letsencrypt/live/reminder-bot.ru/fullchain.pem",
			"/etc/letsencrypt/live/reminder-bot.ru/privkey.pem",
			r,
		)
	}

	if err != nil {
		panic(err)
	}
}

// webAppHandler обработчик, отвечающий за приложение селектора даты напоминания
func (ws *WebServer) webAppHandler(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join("datePicker", "index.html")
	http.ServeFile(w, r, path)
}

// setDate обработчик запроса установки даты напоминания. Зарпос приходит из web app.
func (ws *WebServer) setDate(w http.ResponseWriter, r *http.Request) {
	// парсим пришедшие данные
	payload := struct {
		Date   string `json:"date"`
		TaskId string `json:"taskId"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	taskId, err := strconv.Atoi(payload.TaskId)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newDate, err := time.Parse(time.RFC3339, payload.Date)
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newDate = newDate.UTC()

	// если данные успешно распаршены, добавляем дату напоминанию
	err = ws.tasksRepo.SetDate(taskId, &newDate)
	if err != nil {
		log.Print(err)
	}

	// затем, нужно пройтись по всем сообщениям чата, имеющим предложение установить дату,
	// и у этих сообщений удаляем кнопку установки даты.
	// также, нужно удалить из бд все записи-связи напоминания и сообщений с предложением установить дату
	go func() {
		messageList, err := ws.msgRepo.GetAllByTaskID(taskId)
		if err != nil {
			log.Print(err)
		}
		for _, m := range messageList {
			_, _, err = ws.bot.EditMessageReplyMarkup(&gotgbot.EditMessageReplyMarkupOpts{
				MessageId:   int64(m.MessageID),
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: nil},
				ChatId:      int64(m.ChatID),
			})
			if err != nil {
				log.Print(err)
			}
		}

		err = ws.msgRepo.DeleteAllByTaskId(taskId)
		if err != nil {
			log.Print(err)
		}
	}()

	w.WriteHeader(http.StatusOK)
}
