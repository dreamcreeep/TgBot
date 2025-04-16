// проектируем команды для пользователей бота
package telegram

import (
	"TgBotwWw/clients/telegram"
	"TgBotwWw/lib/e"
	"TgBotwWw/storage"
	"errors"
	"log"
	"net/url"
	"strings"
)

const (
	RndCmd   = "/rnd"
	HelpCmd  = "/help"
	StartCmd = "/start"
)

func NewMessageSender(chatID int, tg *telegram.Client) func(string) error {
	return func(msg string) error {
		return tg.SendMessage(chatID, msg)
	}
}

func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("got new command '%s' from '%s", text, username)

	//проверяем что нам пришел url
	if isAddCmd(text) {
		return p.savePage(chatID, text, username)
	}

	//проверяем тип информации который получили от пользователя
	switch text {
	case RndCmd:
		return p.sendRandom(chatID, username)
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendHello(chatID)
	default:
		return p.tg.SendMessage(chatID, msgUnknownCommand)
	}
}

func (p *Processor) savePage(chatID int, pageURL string, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: save page", err) }()

	sendMsg := NewMessageSender(chatID, p.tg)

	// формируем страницу для сохранения
	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
	}

	isExists, err := p.storage.IsExists(page)
	if err != nil {
		return err
	}

	// если ссылка уже существует, отправляем сообщение
	if isExists {
		return sendMsg(msgAlreadyExists)
		// return p.tg.SendMessage(chatID, msgAlreadyExists)
	}

	// сохраняем страницу
	if err := p.storage.Save(page); err != nil {
		return err
	}

	// отправляем уведомление об успешном сохранении
	if err := sendMsg(msgSaved); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendRandom(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: can't send random", err) }()

	//ищем случайную статью
	page, err := p.storage.PickRandom(username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}

	// обрабатываем особый тип ошибки
	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(chatID, msgNoSavedPages)
	}

	if err := p.tg.SendMessage(chatID, page.URL); err != nil {
		return err
	}

	//если нашли и отправили ссылку, удаляем
	return p.storage.Remove(page)

}

func (p *Processor) sendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, msgHelp)
}

func (p *Processor) sendHello(chatID int) error {
	return p.tg.SendMessage(chatID, msgHello)
}

// проверяем текст на ссылку
func isAddCmd(text string) bool {
	return isURL(text)
}

// парсим URL
func isURL(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Scheme != "" && u.Host != ""
}
