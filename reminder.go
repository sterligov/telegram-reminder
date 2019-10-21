package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	timeLayout          = "2006-01-02 15:04"
	upcomingButtonText  = "Предстоящие"
	completedButtonText = "Завершенные"
	HelloText           = "Привет! Это бот напоминалка. Напиши сообщение и в указанное время он пришлет тебе напоминание. Например: '1955 1510 Купить хлеб'. Напоминание придет в 19:55 15 октября текущего года."
)

type Reminder struct {
	Bot     *tgbotapi.BotAPI
	Storage *PostgreStorage
}

type message struct {
	ID         int
	Txt        string
	ReminderAt string
	CreatedAt  string
	ChatID     int64
}

type eventFunc func(u *tgbotapi.Update)

func (r *Reminder) Start(endpoint string) {
	event := r.initEventFuncs()
	updates := r.Bot.ListenForWebhook(endpoint)
	go r.cron()
	for update := range updates {
		if update.CallbackQuery != nil {
			d := strings.Split(update.CallbackQuery.Data, " ")
			efun := event[d[0]]
			efun(&update)
		} else if update.Message != nil {
			if efun, ok := event[update.Message.Text]; ok {
				efun(&update)
			} else {
				r.createReminder(&update)
			}
		}
	}
}

func (r *Reminder) initEventFuncs() map[string]eventFunc {
	event := make(map[string]eventFunc, 0)
	event["/start"] = eventFunc(r.sendHello)
	event["delete"] = eventFunc(r.delete)
	event["putOff"] = eventFunc(r.putOff)
	event[upcomingButtonText] = eventFunc(r.sendUpcoming)
	event[completedButtonText] = eventFunc(r.sendCompeted)
	return event
}

func (r *Reminder) cron() {
	for {
		time.Sleep(time.Minute)
		go func() {
			messages, err := r.Storage.Current()
			if err != nil {
				log.Println(err)
				return
			}
			for _, m := range messages {
				msg := tgbotapi.NewMessage(m.ChatID, m.Txt)
				reminderID := strconv.Itoa(m.ID)
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Отложить на 15 минут", "putOff "+reminderID+" 15"),
						tgbotapi.NewInlineKeyboardButtonData("Отложить на 30 минут", "putOff "+reminderID+" 30"),
						tgbotapi.NewInlineKeyboardButtonData("Отложить на час", "putOff "+reminderID+" 60"),
					),
				)
				r.Bot.Send(msg)
			}
		}()
	}
}

func (r *Reminder) sendHello(u *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, HelloText)
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(upcomingButtonText),
			tgbotapi.NewKeyboardButton(completedButtonText),
		),
	)
	r.Bot.Send(msg)
}

func (r *Reminder) sendUpcoming(u *tgbotapi.Update) {
	messages, err := r.Storage.Upcoming(u.Message.Chat.ID)
	if err != nil {
		log.Println(err)
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, "Произошла ошибка")
		r.Bot.Send(msg)
		return
	}
	if len(messages) == 0 {
		r.Bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, "Нет предстоящих напоминаний"))
		return
	}
	for _, m := range messages {
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, m.ReminderAt+" "+m.Txt)
		data := strconv.Itoa(m.ID)
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Удалить", "delete "+data),
			),
		)
		r.Bot.Send(msg)
	}
}

func (r *Reminder) sendCompeted(u *tgbotapi.Update) {
	messages, err := r.Storage.Completed(u.Message.Chat.ID)
	if err != nil {
		log.Println(err)
		r.Bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, "Произошла ошибка"))
		return
	}
	if len(messages) == 0 {
		r.Bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, "Нет завершенных напоминаний"))
		return
	}
	for _, m := range messages {
		r.Bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, m.ReminderAt+" "+m.Txt))
	}
}

func (r *Reminder) putOff(u *tgbotapi.Update) {
	data := strings.Split(u.CallbackQuery.Data, " ")
	reminderID, _ := strconv.ParseInt(data[1], 10, 64)
	chatID := u.CallbackQuery.Message.Chat.ID
	putOffMinutes, _ := strconv.ParseInt(data[2], 10, 64)

	m, err := r.Storage.Get(reminderID)
	if err != nil {
		log.Println(err)
		r.Bot.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка"))
		return
	}

	newDate, _ := time.Parse(timeLayout, m.ReminderAt)
	newDate = newDate.Add(time.Minute * time.Duration(putOffMinutes))
	err = r.Storage.Update(reminderID, newDate.Format(timeLayout))
	if err != nil {
		log.Println(err)
		r.Bot.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка"))
		return
	}

	r.Bot.Send(tgbotapi.NewMessage(chatID, "Напоминание отложено"))
}

func (r *Reminder) delete(u *tgbotapi.Update) {
	d := strings.Split(u.CallbackQuery.Data, " ")
	id, _ := strconv.Atoi(d[1])
	err := r.Storage.Delete(int64(id))
	msg := "Напоминание удалено"
	if err != nil {
		log.Println(err)
		msg = "Произошла ошибка"
	}
	r.Bot.Send(tgbotapi.NewMessage(u.CallbackQuery.Message.Chat.ID, msg))
}

func (r *Reminder) createReminder(u *tgbotapi.Update) {
	msg, err := parseMessage(u.Message.Text)
	if err != nil {
		log.Printf("Unknown message %s", u.Message.Text)
		r.Bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, err.Error()))
		return
	}
	msg.ChatID = u.Message.Chat.ID

	var s string
	if err := r.Storage.Save(msg); err != nil {
		log.Println(err)
		s = fmt.Sprintf("Произошла ошибка, напоминание не было сохранено")
	} else {
		s = fmt.Sprintf("Напоминание создано. Вы будете уведомлены %s", msg.ReminderAt)
	}
	r.Bot.Send(tgbotapi.NewMessage(u.Message.Chat.ID, s))
}

func parseMessage(msg string) (*message, error) {
	msg = strings.Trim(msg, " ")
	r := regexp.MustCompile(`^([0-2]\d[0-5]\d)\s+(([0-3]\d)?((0\d)|(1[0-2]))?(\d{2})?\s+)?(.*)`)
	matches := r.FindStringSubmatch(msg)

	if len(matches) == 0 {
		return nil, fmt.Errorf("Неверный формат сообщения")
	}

	today := time.Now()
	date := insRepetitiveVal(matches[1], ":", 2)

	if matches[2] != "" {
		matches[2] = strings.Trim(matches[2], " ")
		matches[2] = reverseDate(matches[2])
		switch len(matches[2]) {
		case 2:
			date = today.Format("2006-01") + "-" + matches[2] + " " + date
		case 4:
			date = today.Format("2006") + "-" + insRepetitiveVal(matches[2], "-", 2) + " " + date
		case 6:
			date = "20" + insRepetitiveVal(matches[2], "-", 2) + " " + date
		}
	} else {
		date = today.Format("2006-01-02") + " " + date
	}

	if err := validDate(date); err != nil {
		return nil, err
	}
	msg = strings.Trim(matches[len(matches)-1], " ")

	return &message{Txt: msg, ReminderAt: date}, nil
}
