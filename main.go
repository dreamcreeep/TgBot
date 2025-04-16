package main

// запуск Consumer'a с помощью Fetcher и Processor
// fetcher и Processor общаются с api tg, первый отправляет запрос чтобы получать новые события, Processor после обработки отправляет нам сообщения
// все это делается с помощью Client

import (
	//"context"
	"flag"
	"log"

	tgClient "TgBotwWw/clients/telegram"
	event_consumer "TgBotwWw/consumer/event-consumer"
	"TgBotwWw/events/telegram"
	"TgBotwWw/storage/files"
	//"TgBotwWw/storage/sqlite"
)

const (
	tgBotHost         = "api.telegram.org"
	sqliteStoragePath = "data/sqlite/storage.db"
	batchSize         = 100
	storagePath       = "storage"
)

func main() {

	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, mustToken()),
		files.New(storagePath),
	)

	log.Print("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)
	}
}

// функция не обрататывает ошибку а выдает панику при неправильном получении токена
func mustToken() string {
	token := flag.String(
		"tg-bot-token",
		"",
		"token for access to telegram bot",
	)

	flag.Parse()

	if *token == "" {
		log.Fatal("token is not specified")
	}

	return *token
}
