package event_consumer

import (
	"log"
	"time"

	"TgBotwWw/events"
)

type Consumer struct {
	fetcher   events.Fetcher
	processor events.Processor
	batchSize int // размер пачки говорит о том, сколько событий мы будем обрабатывать за раз
}

func New(fetcher events.Fetcher, processor events.Processor, batchSize int) Consumer {
	return Consumer{
		fetcher:   fetcher,
		processor: processor,
		batchSize: batchSize,
	}
}

func (c Consumer) Start() error {

	// вечный цикл, который постоянно ждет события и обрабатывает их
	for {
		gotEvents, err := c.fetcher.Fetch(c.batchSize)

		// ошибка могла произойти из за временной ошибки с сетью
		// в fetcher можно встроить механизм retry
		if err != nil {
			log.Printf("[ERR] consumer: %s", err.Error())

			continue
		}

		if len(gotEvents) == 0 {
			time.Sleep(1 * time.Second)

			continue
		}

		if err := c.handleEvents(gotEvents); err != nil {
			log.Print(err)

			continue
		}
	}
}

// обрабатываем возможные собития
// При таком механизме работы функции возможны потери событий, способы решения
//  1. ретраи
//  2. возвращение обратно в хранилище - ненадежный ,но простой
//  3. механизм фоллбека - например сохраняяем не в storage, а в локальный файл, это исключит потерю ивентов по причине поломки сети,
//     можно сохранять фоллбеки не в файл, а прямо в оперативную память внутри нашей программы, но при перезапуске программы все данные потеряются
//  4. подтверждение для fethcer - он не будет делать сдвиг, пока не обнаружит что мы успешно обработали всю пачку или передавать offset самостоятельно
//
// Можно обрабатывать всю пачку раз за разом, но делать это с условием
// 1. останавливаться после первой ошибки
// 2. ввести счетчик ошибок
// Организовать парралельную обработку с помощью wg = Wait.Group{}
func (c *Consumer) handleEvents(events []events.Event) error {

	for _, event := range events {
		log.Printf("got new event: %s", event.Text)

		if err := c.processor.Process(event); err != nil {
			log.Printf("can't handle event: %s", err.Error())

			continue
		}
	}

	return nil
}
