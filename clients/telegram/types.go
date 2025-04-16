package telegram

// определяем все типы с которыми будет работать Сlient
// теги структур нужны для того чтобы парсить json в конкретную структуру,
// апдейты сервера будут приходить в виде json и стандартный парсер по умолчанию будет искать в ответе поле (к примеру ID)

type UpdatesResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	ID int `json:"update_id"`
	//поле Message может отсутствовать, чтобы избежать nil добавляем ссылку на структуру
	Message *IncomingMessage `json:"message"`
}

// IncomingMessage https://core.telegram.org/bots/api#message
type IncomingMessage struct {
	Text string `json:"text"`
	From From   `json:"from"`
	Chat Chat   `json:"chat"`
}
type From struct {
	Username string `json:"username"`
}

type Chat struct {
	ID int `json:"id"`
}
