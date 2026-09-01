package domain

import "errors"

var (
	ErrNotFound     = errors.New("документ не найден")
	ErrDateInFuture = errors.New("нельзя создать запись на будущую дату")
	ErrInvalidDate  = errors.New("неверный формат даты, ожидается YYYY-MM-DD")

	ErrTaskNotFound = errors.New("индекс задачи вне допустимого диапазона")
	ErrNoteNotFound = errors.New("индекс заметки вне допустимого диапазона")
	ErrEmptyTitle   = errors.New("заголовок задачи не может быть пустым")

	ErrInvalidState  = errors.New("некорректное состояние документа")
	ErrCorruptedData = errors.New("данные документа повреждены")
)
