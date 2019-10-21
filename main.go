package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
}

func createBot(token, webHook string) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(webHook))
	if err != nil {
		panic(err)
	}
	return bot
}

func initDatabase() *sql.DB {
	var exists bool
	var username, password, host, schema string

	if username, exists = os.LookupEnv("DB_USERNAME"); !exists {
		panic("You must set the username for database")
	}
	if password, exists = os.LookupEnv("DB_PASSWORD"); !exists {
		panic("You must set the password for database")
	}
	if schema, exists = os.LookupEnv("DB_SCHEMA"); !exists {
		panic("You must set the schema for database")
	}
	if host, exists = os.LookupEnv("DB_HOST"); !exists {
		panic("You must set the host for database")
	}

	db, err := sql.Open(
		"postgres",
		fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, username, password, schema),
	)
	if err != nil {
		panic(err.Error())
	}
	return db
}

func panicRecover() {
	if r := recover(); r != nil {
		log.Println("PANIC", r)
	}
}

func main() {
	defer panicRecover()
	f, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0775)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	log.SetOutput(f)

	bot := createBot(os.Getenv("TG_TOKEN"), os.Getenv("TG_WEBHOOK"))
	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	p := &PostgreStorage{initDatabase()}
	rem := &Reminder{bot, p}
	go rem.Start(os.Getenv("APP_ENDPOINT"))
	s := &http.Server{
		Addr:         ":" + os.Getenv("APP_PORT"),
		Handler:      http.DefaultServeMux,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
	}
	s.ListenAndServe()
}
