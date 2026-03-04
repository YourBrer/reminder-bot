package bot

import (
	"fmt"
	"html"
	"log"
	"reminder/internal/messages"
	"reminder/internal/tasks"
	"reminder/internal/users"
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

type TaskListInterval struct {
	Start time.Time
	End   time.Time
}

const createPrefix = "create"
const cancelPrefix = "cancel"
const todayList = "today"
const weekList = "week"
const monthList = "month"
const allList = "all"

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
var Replacer = strings.NewReplacer(months...)

func (w *Wrapper) Run() {
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
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		panic("Фатальная ошибка при старте поллинга: " + err.Error())
	}

	go w.loop()
	updater.Idle()
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
			listMessageText = "Не удалось получить список задач :-("
		} else if len(taskList) == 0 {
			listMessageText = "Список напоминаний пуст :-)"
		} else if intervalDescription == weekList || intervalDescription == monthList {
			start := Replacer.Replace(interval.Start.Format("2 January"))
			end := Replacer.Replace(interval.End.Format("2 January"))
			listMessageText += fmt.Sprintf("с %s по %s:", start, end)
		}

		_, err = b.SendMessage(chatId, listMessageText, nil)
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
					dateString := Replacer.Replace(task.GetUserTime().Format("2 January в 15:04"))
					sb.WriteString("\n\nНапомню <strong>")
					sb.WriteString(dateString)
					sb.WriteString("</strong>")
				}

				msg, err := b.SendMessage(int64(task.ChatId), sb.String(), opts)
				if err != nil {
					log.Print(err)
				}
				err = w.msgRepo.Create(messages.NewMessage(uint(msg.MessageId), task.ID, task.ChatId))
				if err != nil {
					log.Print(err)
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
			"Создать напоминание из этого текста?",
			&gotgbot.SendMessageOpts{
				ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{
								Text:         "Создать",
								CallbackData: createPrefix,
							},
							{Text: "Отменить", CallbackData: "cancel"},
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
			_, err = b.SendMessage(ctx.EffectiveChat.Id, "Не удалось создать напоминание :-(", nil)
			if err != nil {
				log.Print(err)
			}
		}

		msg, _, err := cq.Message.EditText(
			b,
			"Напоминание создано.",
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
			Text: "Добавить дату и время напоминания",
			WebApp: &gotgbot.WebAppInfo{
				Url: fmt.Sprintf("%s/app?taskId=%d", w.tgWebAppUrl, taskId),
			},
		}}},
	}
}

// Цикл проверки напоминаний
func (w *Wrapper) loop() {
	for {
		start := time.Now()
		end := time.Now().Add(w.taskPollingInterval)
		taskList, err := w.tasksRepo.GetByInterval(start, end)
		if err != nil {
			log.Print(err)
		}
		for _, task := range taskList {
			go func() {
				_, err = w.bot.SendMessage(int64(task.ChatId), task.Description, nil)
				if err != nil {
					log.Print(err)
				}

				err = w.msgRepo.DeleteAllByTaskId(int(task.ID))
				if err != nil {
					log.Print(err)
				}
				err := w.tasksRepo.Delete(task)
				if err != nil {
					log.Print(err)
				}
			}()
		}

		time.Sleep(w.taskPollingInterval)
	}
}

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
