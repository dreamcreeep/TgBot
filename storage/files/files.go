package files

import (
	"TgBotwWw/lib/e"
	"TgBotwWw/storage"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// определяем тип Storage для реализации интерфейса
type Storage struct {
	basePath string
}

// право доступа к файлу, в нашем случае владелец и группа имеют полный доступ, остальные только чтение
const defaultPerm = 0774

// создаем Storage. Передаем путь, возвращаем само хранилище
func New(basePath string) Storage {
	return Storage{basePath: basePath}
}

func (s Storage) Save(page *storage.Page) (err error) {
	//определяем способ обработки ошибок
	defer func() { err = e.WrapIfErr("can't save page", err) }()

	// формируем путь до директории, в которую будет сохраняться файл
	fPath := filepath.Join(s.basePath, page.UserName)

	//создаем директории которые входят в переданный путь
	if err := os.MkdirAll(fPath, defaultPerm); err != nil {

		//из за того что мы обернули ошибку в defer func остается просто вернуть err
		return err
	}

	//формируем имя файла
	fName, err := fileName(page)
	if err != nil {
		return err
	}

	//добавляем имя файла к пути
	fPath = filepath.Join(fPath, fName)

	//создаем файл
	file, err := os.Create(fPath)
	if err != nil {
		return err
	}

	// пишем функцию для того чтобы не обрабатывать ошибку Close()
	defer func() { _ = file.Close() }()

	//серриализуем страницу, приводим к формату который мы можем записать в файл
	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return err
	}

	return nil
}

func (s Storage) PickRandom(userName string) (page *storage.Page, err error) {
	defer func() { err = e.WrapIfErr("can't pick random page", err) }()

	// формируем путь до директории с файлами
	path := filepath.Join(s.basePath, userName)

	// формируем список файлов
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// если файлов нет возвращаем ошибку
	if len(files) == 0 {
		return nil, storage.ErrNoSavedPages
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	n := rand.Intn(len(files))

	//формируем случайный файл с рандомным номером
	file := files[n]

	return s.decodePage(filepath.Join(s.basePath, file.Name()))
}

func (s Storage) Remove(p *storage.Page) (err error) {

	fileName, err := fileName(p)
	if err != nil {
		return e.Wrap("can't remove file", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	if err := os.Remove(path); err != nil {
		msg := fmt.Sprintf("can't remove file: %s", path)

		return e.Wrap(msg, err)
	}

	return nil
}

func (s Storage) IsExists(p *storage.Page) (bool, error) {

	fileName, err := fileName(p)
	if err != nil {
		return false, e.Wrap("can't check if file exists", err)
	}

	path := filepath.Join(s.basePath, p.UserName, fileName)

	// проверяем существует ли файл
	switch _, err = os.Stat(path); {
	// если не нашли (ошибка ErrNotExist), возвращаем false с нулевой ошибкой
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	// обычная ошибка при проверке файла, возвращаем ее наверх
	case err != nil:
		msg := fmt.Sprintf("can't check if file %s exists", path)

		return false, e.Wrap(msg, err)
	}

	return true, nil
}

// функция для декодирования страницы
func (s Storage) decodePage(filePath string) (*storage.Page, error) {
	// открываем файл с defer закрытием
	f, err := os.Open(filePath)
	if err != nil {
		return nil, e.Wrap("can't decode page", err)
	}
	defer func() { _ = f.Close() }()

	//переменная в которую будет декодирован файл
	var p storage.Page

	// декодируем файл
	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, e.Wrap("can't decode page", err)
	}

	return &p, nil
}

// пропишем фуникцию для того чтобы в дальнейшем при изменении способа формирования имен
// не искать все вызовы Hash по коду
func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}
