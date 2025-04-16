// Storage может работать с файловой системой, БД и т.д.
package storage

import (
	"TgBotwWw/lib/e"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
)

type Storage interface {
	Save(p *Page) error
	PickRandom(userName string) (*Page, error)
	Remove(p *Page) error
	IsExists(p *Page) (bool, error)
}

var ErrNoSavedPages = errors.New("no saved pages")

// Page основной тип данных для Storage, страница на которую ведет ссылка, которую мы скинули боту
type Page struct {
	URL      string // сам URL ссылки
	UserName string // пользователь который сылку сохранил, чтобы понимать кому ее отдавать
	//Created time.Time для сортировки страниц по дате загрузки
}

// Hash напишем метод для создания хеша к файлу для того чтобы сохранять его с уникальным именем
// Хэш формируем из URL и UserName
func (p Page) Hash() (string, error) {
	h := sha1.New()

	if _, err := io.WriteString(h, p.URL); err != nil {
		return "", e.Wrap("can't calculate hash", err)
	}
	if _, err := io.WriteString(h, p.UserName); err != nil {
		return "", e.Wrap("can't calculate hash", err)
	}

	// Sum возвращает срез байт, используем Sprintf для преобразования в текст
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
