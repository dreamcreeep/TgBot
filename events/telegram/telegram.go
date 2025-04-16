// пишем реализацию для Event Processor
// в нашем случае этим будет заниматься единственный тип Processor
package telegram

import (
	"TgBotwWw/clients/telegram"
	"TgBotwWw/events"
	"TgBotwWw/lib/e"
	"TgBotwWw/storage"
	"errors"
	"log"
)

type Processor struct {
	tg      *telegram.Client
	offset  int
	storage storage.Storage
}

type Meta struct {
	ChatID   int
	Username string
}

var ErrUnknownEventType = errors.New("unknown event type")
var ErrUnknownMetaType = errors.New("unknown meta type")

// создаем Processor функцией
func New(client *telegram.Client, storage storage.Storage) *Processor {
	return &Processor{
		tg:      client,
		storage: storage,
	}
}

func (p *Processor) Fetch(limit int) ([]events.Event, error) {
	// с помощью клиента получаем все апдейты
	updates, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, e.Wrap("can't get events", err)
	}

	//если апдейтов нет, заканчиваем работу функции и возвращаем nil
	if len(updates) == 0 {
		log.Println("no updates found")
		return nil, nil
	}
	//аллоцируем память под результат
	res := make([]events.Event, 0, len(updates))

	// перебираем апдейты и преобразуем их в тип Event
	for _, u := range updates {
		res = append(res, event(u))
	}

	// обновляем offset, для того чтобы при последующем вызове Fetch получить новую порцию событий
	// при следующем запросе получаем только те апдейты у которых ID больше чем у последнего из уже полученных
	p.offset = updates[len(updates)-1].ID + 1

	return res, nil
}
func (p *Processor) Process(event events.Event) error {
	switch event.Type {
	//если хотим работать не только с сообщениями, но и с другими апдейтами от тг, можно добавить новый case
	case events.Message:
		return p.processMessage(event)
	default:
		return e.Wrap("can't process message", ErrUnknownEventType)
	}
}

func (p *Processor) processMessage(event events.Event) (err error) {
	defer func() { err = e.WrapIfErr("can't process message", err) }()

	// поскольку мы работаем уже с events, для начала нужно получить meta
	meta, err := meta(event)
	if err != nil {
		return err
	}

	if err := p.doCmd(event.Text, meta.ChatID, meta.Username); err != nil {
		return err
	}

	return nil

}

func meta(event events.Event) (Meta, error) {
	//type assertion для Meta
	// если false возвращаем пустую Meta  и ошибку
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("can't get meta", ErrUnknownMetaType)
	}

	return res, nil
}

// функция для преобразования апдейта в event
func event(upd telegram.Update) events.Event {
	updType := fetchType(upd)

	res := events.Event{
		Type: updType,
		Text: fetchText(upd),
	}
	// если тип сообщения Message, добавляем параметр Meta
	if updType == events.Message {
		res.Meta = Meta{
			ChatID:   upd.Message.Chat.ID,
			Username: upd.Message.From.Username,
		}
	}

	return res

}

func fetchText(upd telegram.Update) string {
	//если не проверить Message на null, можем словить панику
	if upd.Message == nil {
		return ""
	}
	//возвращаем текст сообщения
	return upd.Message.Text
}

func fetchType(upd telegram.Update) events.Type {
	//если сообщение нулевое, тип не известен
	if upd.Message == nil {
		return events.UnKnown
	}
	return events.Message
}
