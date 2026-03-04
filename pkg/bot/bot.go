package bot

import (
	"context"
	"fmt"
	"html"
	"log"
	"reminder/internal/messages"
	"reminder/internal/tasks"
	"reminder/internal/users"
	"reminder/pkg/emoji"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
)

// Wrapper Обертка, добавляющая доступы к репозиториям и конфигам
type Wrapper struct {
	tasksRepo           *tasks.Repository
	userRepo            *users.Repository
	msgRepo             *messages.Repository
	tgWebAppUrl         string
	taskPollingInterval time.Duration
	bot                 *gotgbot.Bot
}

// TaskListInterval представляет временной интервал для фильтрации задач
type TaskListInterval struct {
	Start time.Time
	End   time.Time
}

const (
	createPrefix = "create"
	cancelPrefix = "cancel"
	todayList    = "today"
	weekList     = "week"
	monthList    = "month"
	allList      = "all"
)

const (
	pollingTimeout = 9 * time.Second
	requestTimeout = 10 * time.Second
)

var months = []string{
	"January", "января",
	"February", "февраля",
	"March", "марта",
	"April", "апреля",
	"May", "мая",
	"June", "июня",
	"July", "июля",
	"August", "августа",
	"September", "сентября",
	"October", "октября",
	"November", "ноября",
	"December", "декабря",
}
var replacer = strings.NewReplacer(months...)

// Run запускает бота и начинает обработку обновлений
// Функция блокирует выполнение до получения сигнала остановки через context
func (w *Wrapper) Run(ctx context.Context) {
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("Ошибка при обновлении:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dispatcher, nil)

	dispatcher.AddHandler(handlers.NewCommand(allList, w.taskListHandler(allList)))
	dispatcher.AddHandler(handlers.NewCommand(todayList, w.taskListHandler(todayList)))
	dispatcher.AddHandler(handlers.NewCommand(weekList, w.taskListHandler(weekList)))
	dispatcher.AddHandler(handlers.NewCommand(monthList, w.taskListHandler(monthList)))
	dispatcher.AddHandler(handlers.NewCallback(
		func(cq *gotgbot.CallbackQuery) bool {
			for _, prefix := range []string{createPrefix, cancelPrefix} {
				if strings.HasPrefix(cq.Data, prefix) {
					return true
				}
			}
			return false
		},
		w.handleCallback,
	))
	dispatcher.AddHandler(handlers.NewMessage(message.Text, w.userMessageHandler()))

	err := updater.StartPolling(w.bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout:     int64(pollingTimeout.Seconds()),
			RequestOpts: &gotgbot.RequestOpts{Timeout: requestTimeout},
		},
	})
	if err != nil {
		log.Printf("Фатальная ошибка при старте поллинга: %v", err)
		return
	}

	go w.loop(ctx)

	// Ожидаем сигнал остановки
	<-ctx.Done()
	log.Println("Останавливаем бота...")
	updater.Stop()
	log.Println("Бот остановлен")
}

// Обработчик команд списка напоминаний по интервалу
func (w *Wrapper) taskListHandler(intervalDescription string) handlers.Response {
	return func(b *gotgbot.Bot, ctx *ext.Context) error {
		var interval TaskListInterval
		var listMessageText string
		now := time.Now()
		switch intervalDescription {
		case todayList:
			interval = TaskListInterval{
				Start: now,
				End:   time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()),
			}
			listMessageText = "Список напоминаний на сегодня:"
		case weekList:
			interval = TaskListInterval{
				Start: now,
				End:   now.AddDate(0, 0, 7),
			}
			listMessageText = "Список напоминаний на неделю, "
		case monthList:
			interval = TaskListInterval{
				Start: now,
				End:   now.AddDate(0, 1, 0),
			}
			listMessageText = "Список напоминаний на месяц, "
		default:
			interval = TaskListInterval{}
			listMessageText = "Список напоминаний:"
		}

		chatId := ctx.EffectiveChat.Id
		var taskList []*tasks.Task
		var err error

		if interval.Start.IsZero() && interval.End.IsZero() {
			taskList, err = w.tasksRepo.GetAllByUserId(ctx.EffectiveUser.Id)
		} else {
			taskList, err = w.tasksRepo.GetByUserAndInterval(ctx.EffectiveUser.Id, interval.Start, interval.End)
		}

		if err != nil {
			listMessageText = "Не удалось получить список задач"
		} else if len(taskList) == 0 {
			listMessageText = "Список напоминаний пуст"
		} else if intervalDescription == weekList || intervalDescription == monthList {
			start := replacer.Replace(interval.Start.Format("2 January"))
			end := replacer.Replace(interval.End.Format("2 January"))
			listMessageText += fmt.Sprintf("с %s по %s:", start, end)
		}

		_, err = b.SendMessage(chatId, emoji.AddRandomEmojiToText(listMessageText), nil)
		if err != nil {
			log.Println("Ошибка отправки сообщения:", err.Error())
		}

		if len(taskList) > 0 {
			for _, task := range taskList {
				opts := &gotgbot.SendMessageOpts{ParseMode: "html"}
				if task.ExecutionDate == nil {
					opts.ReplyMarkup = w.getDateButton(task.ID)
				}

				sb := strings.Builder{}
				sb.WriteString(task.Description)
				if task.ExecutionDate != nil {
					userTime, err := task.GetUserTime()
					if err != nil {
						log.Printf("Ошибка получения времени пользователя: %v", err)
						userTime = task.ExecutionDate
					}
					dateString := replacer.Replace(userTime.Format("2 January в 15:04"))
					sb.WriteString("\n\nНапомню <strong>")
					sb.WriteString(dateString)
					sb.WriteString("</strong>")
				}

				msg, err := b.SendMessage(
					int64(task.ChatId),
					emoji.AddRandomEmojiToText(sb.String()),
					opts,
				)
				if err != nil {
					log.Printf("Ошибка отправки сообщения: %v", err)
					continue
				}
				err = w.msgRepo.Create(messages.NewMessage(uint(msg.MessageId), task.ID, task.ChatId))
				if err != nil {
					log.Printf("Ошибка создания сообщения: %v", err)
				}
			}
		}

		return nil
	}
}

// Обработчик текстовых сообщений пользователя. Предлагает создать напоминание из присланного
// пользователем текста
func (w *Wrapper) userMessageHandler() handlers.Response {
	return func(b *gotgbot.Bot, ctx *ext.Context) error {
		_, err := ctx.EffectiveMessage.Reply(
			b,
			emoji.AddRandomEmojiToText("Создать напоминание из этого текста?"),
			&gotgbot.SendMessageOpts{
				ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{
								Text:         emoji.AddRandomEmojiToText("Создать"),
								CallbackData: createPrefix,
							},
							{Text: emoji.AddRandomEmojiToText("Отменить"), CallbackData: "cancel"},
						},
					},
				},
			},
		)
		return err
	}
}

// Обработчик нажатий кнопок
func (w *Wrapper) handleCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	cq := ctx.Update.CallbackQuery

	// Отменить создание напоминания
	if strings.HasPrefix(cq.Data, cancelPrefix) {
		_, err := cq.Message.Delete(b, nil)
		if err != nil {
			log.Print(err)
		}
		return nil
	}

	// Создать напоминание
	if strings.HasPrefix(cq.Data, createPrefix) {
		task := tasks.NewTask(
			html.EscapeString(ctx.EffectiveMessage.ReplyToMessage.Text),
			uint(ctx.EffectiveUser.Id),
			uint(ctx.EffectiveChat.Id),
		)
		_, err := w.tasksRepo.Create(task)
		if err != nil {
			_, err = b.SendMessage(
				ctx.EffectiveChat.Id,
				emoji.AddRandomEmojiToText("Не удалось создать напоминание"),
				nil,
			)
			if err != nil {
				log.Print(err)
			}
		}

		msg, _, err := cq.Message.EditText(
			b,
			emoji.AddRandomEmojiToText("Напоминание создано."),
			&gotgbot.EditMessageTextOpts{
				ReplyMarkup: w.getDateButton(task.ID),
			},
		)

		if err != nil {
			fmt.Println(err.Error())
		}
		if msg != nil {
			err = w.msgRepo.Create(messages.NewMessage(uint(msg.MessageId), task.ID, task.ChatId))
			if err != nil {
				log.Print(err)
			}
		}

		return nil
	}

	return nil
}

// Возвращает шаблон для кнопки добавления даты напоминания
func (w *Wrapper) getDateButton(taskId uint) gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{{
			Text: emoji.AddRandomEmojiToText("Добавить дату и время напоминания"),
			WebApp: &gotgbot.WebAppInfo{
				Url: fmt.Sprintf("%s/app?taskId=%d", w.tgWebAppUrl, taskId),
			},
		}}},
	}
}

// Цикл проверки напоминаний
func (w *Wrapper) loop(ctx context.Context) {
	ticker := time.NewTicker(w.taskPollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Останавливаем цикл проверки напоминаний")
			return
		case <-ticker.C:
			start := time.Now()
			end := start.Add(w.taskPollingInterval)

			taskList, err := w.tasksRepo.GetByInterval(start, end)
			if err != nil {
				log.Printf("Ошибка получения задач: %v", err)
				continue
			}

			for _, task := range taskList {
				go func(t *tasks.Task) {
					_, err := w.bot.SendMessage(int64(t.ChatId), t.Description, nil)
					if err != nil {
						log.Printf("Ошибка отправки сообщения: %v", err)
					}

					err = w.msgRepo.DeleteAllByTaskId(int(t.ID))
					if err != nil {
						log.Printf("Ошибка удаления сообщений: %v", err)
					}

					err = w.tasksRepo.Delete(t)
					if err != nil {
						log.Printf("Ошибка удаления задачи: %v", err)
					}
				}(task)
			}
		}
	}
}

// NewWrapper создает новый экземпляр Wrapper с заданными зависимостями
func NewWrapper(
	bot *gotgbot.Bot,
	tr *tasks.Repository,
	ur *users.Repository,
	url string,
	pi time.Duration,
	mr *messages.Repository,
) *Wrapper {
	return &Wrapper{
		bot:                 bot,
		tasksRepo:           tr,
		userRepo:            ur,
		tgWebAppUrl:         url,
		taskPollingInterval: pi,
		msgRepo:             mr,
	}
}
