// описываем интферфейсы для Event Processor
package events

// Fetcher принимает пачку событий исходя из limit
type Fetcher interface {
	Fetch(limit int) ([]Event, error)
}

// Processor обрабатывает полученные данные ( и готовит к записи в бд?)
type Processor interface {
	Process(e Event) error
}

// кастомный тип, исключает ошибки на этапе компиляции, удобно искать по проекту
type Type int

const (
	UnKnown Type = iota
	Message
)

type Event struct {
	Type Type
	Text string
	Meta interface{}
}
