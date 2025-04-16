package telegram

// пишем CLient для общения с API бота: получение Updates и отправка собственных сообщений пользователям SendMessage

import (
	"TgBotwWw/lib/e"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

// хост API сервиса  tg, префикс с которого начинаются все запросы (tg-bot.com/bot<token>) и http клиент
type Client struct {
	host     string
	basePath string
	client   http.Client
}

const (
	getUpdatesMethod  = "getUpdates"
	sendMessageMethod = "sendMessage"
)

func New(host string, token string) *Client {
	return &Client{
		host:     host,
		basePath: NewBasePath(token),
		client:   http.Client{},
	}
}

// генерация basePath для tg
func NewBasePath(token string) string {
	return "bot" + token
}

// offset(смещение) для того чтобы не запрашивать все апдейты с 1-го, а, например с 100-го
// limit = количество апдейтов за 1 запрос
func (c *Client) Updates(offset int, limit int) (updates []Update, err error) {
	func() { err = e.WrapIfErr("can't get updates", err) }()

	q := url.Values{}

	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))

	data, err := c.doRequest(getUpdatesMethod, q)
	if err != nil {
		return nil, err
	}

	// результат парсинга json сохраняем в переменную
	var res UpdatesResponse

	// парсим data в res c указателем!
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	return res.Result, nil
}

// обращаемся к документации и ищем метод sendMessage
// chatID для  того чтобы понимать, куда конкретно отправляются сообщения
func (c *Client) SendMessage(chatID int, text string) error {

	//подготавливаем параметры запросов
	q := url.Values{}
	q.Add("chatID", strconv.Itoa(chatID))
	q.Add("text", text)

	//выполняем запрос, тело ответа тут не понадобится
	_, err := c.doRequest(sendMessageMethod, q)
	if err != nil {
		//используем e.Wrap() потому что err !=nil
		return e.Wrap("can't send message", err)
	}

	return nil
}

// функция для отправки запросов
func (c *Client) doRequest(method string, query url.Values) (data []byte, err error) {
	defer func() { err = e.WrapIfErr("can't do request", err) }()

	// формируем URL на который будет отправляться запрос
	u := url.URL{
		Scheme: "https",                       // протокол
		Host:   c.host,                        // хост
		Path:   path.Join(c.basePath, method), // базовый путь из Clientа + метод getUpdates из тг
	}

	// подготовка запроса
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	//передаем в запрос параметры которые получили в аргументе query
	// Encode приводит параметры к виду, которые в последствии можно будет отправить на сервер
	req.URL.RawQuery = query.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	// не забываем! закрыть тело ответа
	defer func() { _ = resp.Body.Close() }()

	// сохраняем содержимое в переменную
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
