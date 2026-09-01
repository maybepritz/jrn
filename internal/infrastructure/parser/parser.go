package parser

import (
	"fmt"
	"jrn/internal/domain"
)

type State uint8

type Parser struct {
	state State
	doc   domain.Document

	// Буферы
	buf         []byte
	streakValue int
	taskValue   int

	attrKey string

	curTask domain.Task
	hasTask bool
}

const (
	START State = iota
	S1
	S2
	S3
	IN_FRONTMATTER
	SEEN_D
	SEEN_DA
	SEEN_DAT
	SEEN_DATE
	SEEN_S
	SEEN_ST
	SEEN_STR
	SEEN_STRE
	SEEN_STREA
	SEEN_STREAK
	SEEN_T
	SEEN_TA
	SEEN_TAS
	SEEN_TASK
	SEEN_TASK_C
	SEEN_TASK_CO
	SEEN_TASK_COM
	SEEN_TASK_COMP
	SEEN_TASK_COMPL
	SEEN_TASK_COMPLE
	SEEN_TASK_COMPLET
	SEEN_TASK_COMPLETE
	SEEN_TASK_COMPLETED
	SEEN_TASK_COMPLETED_COLON
	WAIT_DATE_VALUE
	WAIT_STREAK_VALUE
	WAIT_TOTAL_COMPLETED_VALUE
	READ_DATE
	READ_STREAK
	READ_TOTAL_COMPLETED
	FM_DASH_1
	FM_DASH_2
	FM_DASH_3
	WAIT_SECTION
	SEEN_H1
	SEEN_H2
	SEEN_H3
	WAIT_HEADER_NAME
	SEEN_HEAD_T
	SEEN_HEAD_TA
	SEEN_HEAD_TAS
	SEEN_HEAD_TASK
	SEEN_HEAD_TASKS
	SEEN_HEAD_D
	SEEN_HEAD_DA
	SEEN_HEAD_DAI
	SEEN_HEAD_DAIL
	SEEN_HEAD_DAILY
	SEEN_HEAD_DAILY_SPACE
	SEEN_HEAD_DAILY_L
	SEEN_HEAD_DAILY_LO
	SEEN_HEAD_DAILY_LOG
	IN_TASKS
	IN_DAILY_LOG
	TASK_DASH
	TASK_INDENT
	TASK_DASH_SPACE
	TASK_LBRACKET
	TASK_CHECKBOX
	TASK_RBRACKET
	NOTE_SEEN_DASH
	READ_NOTE_TEXT
	READ_TASK_TEXT
	READ_TASK_TAG
	READ_ATTR_KEY
	READ_ATTR_VAL
	READ_ATTR_WAIT_SPACE
	READ_TASK_TAG_WAIT_NEXT

	STATE_ERROR
)

func Parse(data []byte) (*domain.Document, error) {

	p := Parser{
		state: START,
		buf:   make([]byte, 0, 128), // Предвыделяем буфер под строки
	}

	p.doc.Tasks = make([]domain.Task, 0, 16)
	p.doc.DailyLog = make([]string, 0, 8)

	for i := 0; i < len(data); i++ {
		ch := data[i]

		switch p.state {
		case START:
			if ch == '-' {
				p.state = S1
			} else if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
				// Игнорируем пробелы и переносы строк в начале
			} else {
				return nil, p.syntaxError("ожидался символ '-' в начале файла", ch, i)
			}
		case S1:
			if ch == '-' {
				p.state = S2
			} else {
				return nil, p.syntaxError("ожидался второй символ '-' в начале файла", ch, i)
			}
		case S2:
			if ch == '-' {
				p.state = S3
			} else {
				return nil, p.syntaxError("ожидался третий символ '-' в начале файла", ch, i)
			}
		case S3:
			if ch == '\n' {
				p.state = IN_FRONTMATTER
			} else if ch == ' ' || ch == '\t' || ch == '\r' {
			} else {
				return nil, p.syntaxError("ожидался перенос строки после '---'", ch, i)
			}
		case IN_FRONTMATTER:
			if ch == 'd' {
				p.state = SEEN_D
			} else if ch == 's' {
				p.state = SEEN_S
			} else if ch == 't' {
				p.state = SEEN_T
			} else if ch == '-' {
				p.state = FM_DASH_1
			} else if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
				// Игнорируем пробелы и переносы строк внутри фронтматтера
			} else {
				return nil, p.syntaxError("ожидался ключ 'date', 'streak' или 'total_completed'", ch, i)
			}
		case SEEN_D:
			if ch == 'a' {
				p.state = SEEN_DA
			} else {
				return nil, p.syntaxError("ожидался символ 'a' после 'd'", ch, i)
			}
		case SEEN_DA:
			if ch == 't' {
				p.state = SEEN_DAT
			} else {
				return nil, p.syntaxError("ожидался символ 't' после 'da'", ch, i)
			}
		case SEEN_DAT:
			if ch == 'e' {
				p.state = SEEN_DATE
			} else {
				return nil, p.syntaxError("ожидался символ 'e' после 'dat'", ch, i)
			}
		case SEEN_DATE:
			if ch == ':' {
				p.state = WAIT_DATE_VALUE
			} else if ch == ' ' || ch == '\t' {
				// Игнорируем пробелы перед двоеточием
			} else {
				return nil, p.syntaxError("ожидался символ ':' после ключа 'date'", ch, i)
			}
		case WAIT_DATE_VALUE:
			if ch == ' ' || ch == '\t' {
				// Игнорируем пробелы перед значением
			} else if ch == '\n' {
				return nil, p.syntaxError("ожидалось значение после ключа 'date'", ch, i)
			} else if isDigit(ch) {
				p.buf = append(p.buf, ch)
				p.state = READ_DATE
			} else {
				return nil, p.syntaxError("ожидалось значение после ключа 'date'", ch, i)
			}
		case READ_DATE:
			if ch == '\n' || ch == '\r' {
				p.doc.Meta.Date = string(p.buf)
				p.buf = p.buf[:0]
				p.state = IN_FRONTMATTER
			} else if isDigit(ch) || ch == '-' {
				p.buf = append(p.buf, ch)
			} else {
				return nil, p.syntaxError("ожидалось значение после ключа 'date'", ch, i)
			}
		case SEEN_S:
			if ch == 't' {
				p.state = SEEN_ST
			} else {
				return nil, p.syntaxError("ожидался символ 't' после 's'", ch, i)
			}
		case SEEN_ST:
			if ch == 'r' {
				p.state = SEEN_STR
			} else {
				return nil, p.syntaxError("ожидался символ 'r' после 'st'", ch, i)
			}
		case SEEN_STR:
			if ch == 'e' {
				p.state = SEEN_STRE
			} else {
				return nil, p.syntaxError("ожидался символ 'e' после 'str'", ch, i)
			}
		case SEEN_STRE:
			if ch == 'a' {
				p.state = SEEN_STREA
			} else {
				return nil, p.syntaxError("ожидался символ 'a' после 'stre'", ch, i)
			}
		case SEEN_STREA:
			if ch == 'k' {
				p.state = SEEN_STREAK
			} else {
				return nil, p.syntaxError("ожидался символ 'k' после 'strea'", ch, i)
			}
		case SEEN_STREAK:
			if ch == ':' {
				p.state = WAIT_STREAK_VALUE
			} else if ch == ' ' || ch == '\t' {
				// Игнорируем пробелы перед двоеточием
			} else {
				return nil, p.syntaxError("ожидался символ ':' после ключа 'streak'", ch, i)
			}
		case WAIT_STREAK_VALUE:
			if ch == ' ' || ch == '\t' {
				// Игнорируем пробелы перед значением
			} else if ch == '\n' {
				return nil, p.syntaxError("ожидалось значение после ключа 'streak'", ch, i)
			} else if isDigit(ch) {
				p.streakValue = p.streakValue*10 + int(ch-'0')
				p.state = READ_STREAK
			} else {
				return nil, p.syntaxError("ожидалось числовое значение после ключа 'streak'", ch, i)
			}
		case READ_STREAK:
			if ch == '\n' || ch == '\r' {
				p.doc.Meta.Streak = p.streakValue
				p.streakValue = 0
				p.state = IN_FRONTMATTER
			} else if isDigit(ch) {
				p.streakValue = p.streakValue*10 + int(ch-'0')
			} else {
				return nil, p.syntaxError("ожидалось числовое значение после ключа 'streak'", ch, i)
			}
		case SEEN_T:
			if ch == 'a' {
				p.state = SEEN_TA
			} else {
				return nil, p.syntaxError("ожидался символ 'a' после 't'", ch, i)
			}
		case SEEN_TA:
			if ch == 's' {
				p.state = SEEN_TAS
			} else {
				return nil, p.syntaxError("ожидался символ 's' после 'ta'", ch, i)
			}
		case SEEN_TAS:
			if ch == 'k' {
				p.state = SEEN_TASK
			} else {
				return nil, p.syntaxError("ожидался символ 'k' после 'tas'", ch, i)
			}
		case SEEN_TASK:
			if ch == '_' {
				p.state = SEEN_TASK_C
			} else {
				return nil, p.syntaxError("ожидался символ '_' после 'task'", ch, i)
			}
		case SEEN_TASK_C:
			if ch == 'c' {
				p.state = SEEN_TASK_CO
			} else {
				return nil, p.syntaxError("ожидался символ 'c' после 'task_'", ch, i)
			}
		case SEEN_TASK_CO:
			if ch == 'o' {
				p.state = SEEN_TASK_COM
			} else {
				return nil, p.syntaxError("ожидался символ 'o' после 'task_c'", ch, i)
			}
		case SEEN_TASK_COM:
			if ch == 'm' {
				p.state = SEEN_TASK_COMP
			} else {
				return nil, p.syntaxError("ожидался символ 'm' после 'task_co'", ch, i)
			}
		case SEEN_TASK_COMP:
			if ch == 'p' {
				p.state = SEEN_TASK_COMPL
			} else {
				return nil, p.syntaxError("ожидался символ 'p' после 'task_com'", ch, i)
			}
		case SEEN_TASK_COMPL:
			if ch == 'l' {
				p.state = SEEN_TASK_COMPLE
			} else {
				return nil, p.syntaxError("ожидался символ 'l' после 'task_comp'", ch, i)
			}
		case SEEN_TASK_COMPLE:
			if ch == 'e' {
				p.state = SEEN_TASK_COMPLET
			} else {
				return nil, p.syntaxError("ожидался символ 'e' после 'task_compl'", ch, i)
			}
		case SEEN_TASK_COMPLET:
			if ch == 't' {
				p.state = SEEN_TASK_COMPLETE
			} else {
				return nil, p.syntaxError("ожидался символ 't' после 'task_comple'", ch, i)
			}
		case SEEN_TASK_COMPLETE:
			if ch == 'e' {
				p.state = SEEN_TASK_COMPLETED // переход на ожидание буквы 'd'
			} else {
				return nil, p.syntaxError("ожидался символ 'e' после 'task_complet'", ch, i)
			}
		case SEEN_TASK_COMPLETED: // или переименуй в цепочке
			if ch == 'd' {
				p.state = SEEN_TASK_COMPLETED_COLON
			} else {
				return nil, p.syntaxError("ожидался символ 'd' после 'task_complete'", ch, i)
			}
		case SEEN_TASK_COMPLETED_COLON:
			if ch == ':' {
				p.state = WAIT_TOTAL_COMPLETED_VALUE
			} else if ch == ' ' || ch == '\t' {
				// Игнорируем пробелы перед двоеточием
			} else {
				return nil, p.syntaxError("ожидался символ ':' после ключа 'task_completed'", ch, i)
			}
		case WAIT_TOTAL_COMPLETED_VALUE:
			if ch == ' ' || ch == '\t' {
				// Игнорируем пробелы перед значением
			} else if ch == '\n' {
				return nil, p.syntaxError("ожидалось значение после ключа 'task_completed'", ch, i)
			} else if isDigit(ch) {
				p.taskValue = p.taskValue*10 + int(ch-'0')
				p.state = READ_TOTAL_COMPLETED
			} else {
				return nil, p.syntaxError("ожидалось числовое значение после ключа 'task_completed'", ch, i)
			}
		case READ_TOTAL_COMPLETED:
			if ch == '\n' || ch == '\r' {
				p.doc.Meta.TaskCompleted = p.taskValue
				p.taskValue = 0
				p.state = IN_FRONTMATTER
			} else if isDigit(ch) {
				p.taskValue = p.taskValue*10 + int(ch-'0')
			} else {
				return nil, p.syntaxError("ожидалось числовое значение после ключа 'task_completed'", ch, i)
			}
		case FM_DASH_1:
			if ch == '-' {
				p.state = FM_DASH_2
			} else {
				return nil, p.syntaxError("ожидался второй символ '-' в конце фронтматтера", ch, i)
			}
		case FM_DASH_2:
			if ch == '-' {
				p.state = FM_DASH_3
			} else {
				return nil, p.syntaxError("ожидался третий символ '-' в конце фронтматтера", ch, i)
			}
		case FM_DASH_3:
			if ch == '\n' {
				p.state = WAIT_SECTION
			} else if ch == ' ' || ch == '\t' || ch == '\r' {
				// Игнорируем пробелы и табуляции после '---'
			} else {
				return nil, p.syntaxError("ожидался перенос строки после '---'", ch, i)
			}
		case WAIT_SECTION:
			if ch == '#' {
				p.state = SEEN_H1
			} else if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
				// Игнорируем пробелы и переносы строк перед секцией
			} else {
				return nil, p.syntaxError("ожидался символ '#' для начала секции", ch, i)
			}
		case SEEN_H1:
			if ch == '#' {
				p.state = SEEN_H2
			} else {
				return nil, p.syntaxError("ожидался второй символ '#' для начала секции", ch, i)
			}
		case SEEN_H2:
			if ch == '#' {
				p.state = SEEN_H3
			} else {
				return nil, p.syntaxError("ожидался третий символ '#' для начала секции", ch, i)
			}
		case SEEN_H3:
			if ch == ' ' || ch == '\t' {
				p.state = WAIT_HEADER_NAME
			} else {
				return nil, p.syntaxError("ожидался пробел после '###'", ch, i)
			}
		case WAIT_HEADER_NAME:
			if ch == ' ' || ch == '\t' {
				// Игнорируем пробелы перед названием секции
			} else if ch == 't' || ch == 'T' {
				p.state = SEEN_HEAD_T
			} else if ch == 'd' || ch == 'D' {
				p.state = SEEN_HEAD_D
			} else {
				return nil, p.syntaxError("ожидалось название секции 'Tasks' или 'Daily Log'", ch, i)
			}
		case SEEN_HEAD_T:
			if ch == 'a' {
				p.state = SEEN_HEAD_TA
			} else {
				return nil, p.syntaxError("ожидался символ 'a' после 't'", ch, i)
			}
		case SEEN_HEAD_TA:
			if ch == 's' {
				p.state = SEEN_HEAD_TAS
			} else {
				return nil, p.syntaxError("ожидался символ 's' после 'ta'", ch, i)
			}
		case SEEN_HEAD_TAS:
			if ch == 'k' {
				p.state = SEEN_HEAD_TASK
			} else {
				return nil, p.syntaxError("ожидался символ 'k' после 'tas'", ch, i)
			}
		case SEEN_HEAD_TASK:
			if ch == 's' {
				p.state = SEEN_HEAD_TASKS
			} else {
				return nil, p.syntaxError("ожидался символ 's' после 'task'", ch, i)
			}
		case SEEN_HEAD_TASKS:
			if ch == '\n' {
				p.state = IN_TASKS
			} else if ch == ' ' || ch == '\t' {
				// Игнорируем пробелы после 'Tasks'
			} else {
				return nil, p.syntaxError("ожидался перенос строки после секции 'Tasks'", ch, i)
			}
		case SEEN_HEAD_D:
			if ch == 'a' {
				p.state = SEEN_HEAD_DA
			} else {
				return nil, p.syntaxError("ожидался символ 'a' после 'd'", ch, i)
			}
		case SEEN_HEAD_DA:
			if ch == 'i' {
				p.state = SEEN_HEAD_DAI
			} else {
				return nil, p.syntaxError("ожидался символ 'i' после 'da'", ch, i)
			}
		case SEEN_HEAD_DAI:
			if ch == 'l' {
				p.state = SEEN_HEAD_DAIL
			} else {
				return nil, p.syntaxError("ожидался символ 'l' после 'dai'", ch, i)
			}
		case SEEN_HEAD_DAIL:
			if ch == 'y' {
				p.state = SEEN_HEAD_DAILY
			} else {
				return nil, p.syntaxError("ожидался символ 'y' после 'dail'", ch, i)
			}
		case SEEN_HEAD_DAILY:
			if ch == ' ' {
				p.state = SEEN_HEAD_DAILY_SPACE
			} else {
				return nil, p.syntaxError("ожидался пробел после 'Daily'", ch, i)
			}
		case SEEN_HEAD_DAILY_SPACE:
			if ch == 'l' || ch == 'L' {
				p.state = SEEN_HEAD_DAILY_L
			} else {
				return nil, p.syntaxError("ожидался символ 'l' после 'Daily '", ch, i)
			}
		case SEEN_HEAD_DAILY_L:
			if ch == 'o' {
				p.state = SEEN_HEAD_DAILY_LO
			} else {
				return nil, p.syntaxError("ожидался символ 'o' после 'Daily l'", ch, i)
			}
		case SEEN_HEAD_DAILY_LO:
			if ch == 'g' {
				p.state = SEEN_HEAD_DAILY_LOG
			} else {
				return nil, p.syntaxError("ожидался символ 'g' после 'Daily lo'", ch, i)
			}
		case SEEN_HEAD_DAILY_LOG:
			if ch == '\n' {
				p.state = IN_DAILY_LOG
			} else if ch == ' ' || ch == '\t' {
				// Игнорируем пробелы после 'Daily Log'
			} else {
				return nil, p.syntaxError("ожидался перенос строки после секции 'Daily Log'", ch, i)
			}
		case IN_TASKS:
			if ch == '\n' || ch == '\r' {
				// Обрабатываем перенос строки
			} else if ch == '-' {
				p.state = TASK_DASH
			} else if ch == ' ' || ch == '\t' {
				p.state = TASK_INDENT
			} else if ch == '#' {
				p.saveCurrentTask()
				p.state = SEEN_H1
			} else {
				return nil, p.syntaxError("ожидался символ '-' для начала задачи или пробел/табуляция для отступа", ch, i)
			}
		case TASK_INDENT:
			if ch == ' ' || ch == '\t' {
			} else if ch == '-' {
				p.state = NOTE_SEEN_DASH
			} else {
				return nil, p.syntaxError("ожидался пробел после отступа", ch, i)
			}
		case NOTE_SEEN_DASH:
			if ch == ' ' || ch == '\t' {
				p.buf = p.buf[:0] // Очищаем буфер перед сбором текста заметки
				p.state = READ_NOTE_TEXT
			} else {
				return nil, p.syntaxError("ожидался пробел после '-' в заметке", ch, i)
			}
		case READ_NOTE_TEXT:
			if ch == '\n' || ch == '\r' {
				if len(p.buf) > 0 {
					p.curTask.Notes = append(p.curTask.Notes, trimTrailingSpaces(string(p.buf)))
					p.buf = p.buf[:0]
				}
				p.state = IN_TASKS
			} else {
				p.buf = append(p.buf, ch)
			}
		case TASK_DASH:
			if ch == ' ' {
				p.state = TASK_DASH_SPACE
			} else {
				return nil, p.syntaxError("ожидался пробел после '-' в задаче", ch, i)
			}
		case TASK_DASH_SPACE:
			if ch == '[' {
				p.state = TASK_LBRACKET
			} else {
				return nil, p.syntaxError("ожидался символ '[' чекбокса", ch, i)
			}
		case TASK_LBRACKET:
			if ch == ' ' {
				p.saveCurrentTask()
				p.curTask = domain.Task{Done: false}
				p.hasTask = true
				p.state = TASK_CHECKBOX
			} else if ch == 'x' || ch == 'X' {
				p.saveCurrentTask()
				p.curTask = domain.Task{Done: true}
				p.hasTask = true
				p.state = TASK_CHECKBOX
			} else {
				return nil, p.syntaxError("ожидался пробел или 'x' внутри чекбокса [ ]", ch, i)
			}
		case TASK_CHECKBOX:
			if ch == ']' {
				p.state = TASK_RBRACKET
			} else {
				return nil, p.syntaxError("ожидался закрывающий символ ']' чекбокса", ch, i)
			}
		case TASK_RBRACKET:
			if ch == ' ' {
				//тут можно очистить буфер ес чоу
				p.state = READ_TASK_TEXT
			} else {
				return nil, p.syntaxError("ожидался пробел после чекбокса", ch, i)
			}
		case READ_TASK_TEXT:
			if ch == '\n' || ch == '\r' {
				p.curTask.Title = trimTrailingSpaces(string(p.buf))
				p.buf = p.buf[:0]
				p.state = IN_TASKS
			} else if ch == '#' {
				if p.curTask.Title == "" {
					p.curTask.Title = trimTrailingSpaces(string(p.buf))
				}
				p.buf = p.buf[:0]
				p.state = READ_TASK_TAG
			} else if ch == '@' {
				if p.curTask.Title == "" {
					p.curTask.Title = trimTrailingSpaces(string(p.buf))
				}
				p.buf = p.buf[:0]
				p.state = READ_ATTR_KEY
			} else {
				p.buf = append(p.buf, ch)
			}
		case READ_TASK_TAG:
			if ch == ' ' || ch == '\t' {
				if len(p.buf) > 0 {
					p.curTask.Tags = append(p.curTask.Tags, string(p.buf))
					p.buf = p.buf[:0]
				}
				p.state = READ_TASK_TAG_WAIT_NEXT
			} else if ch == '\n' || ch == '\r' {
				if len(p.buf) > 0 {
					p.curTask.Tags = append(p.curTask.Tags, string(p.buf))
					p.buf = p.buf[:0]
				}
				p.state = IN_TASKS
			} else {
				p.buf = append(p.buf, ch)
			}
		case READ_TASK_TAG_WAIT_NEXT:
			if ch == ' ' || ch == '\t' {

			} else if ch == '#' {
				p.buf = p.buf[:0]
				p.state = READ_TASK_TAG
			} else if ch == '@' {
				p.buf = p.buf[:0]
				p.state = READ_ATTR_KEY
			} else if ch == '\n' || ch == '\r' {
				p.state = IN_TASKS
			} else {
				return nil, p.syntaxError("после тегов допустимы только другие теги (#), атрибуты (@) или перенос строки", ch, i)
			}
		case READ_ATTR_KEY:
			if ch == '(' {
				p.attrKey = string(p.buf)
				p.buf = p.buf[:0]
				p.state = READ_ATTR_VAL
			} else if ch == ' ' || ch == '\t' || ch == '\n' {
				return nil, p.syntaxError("ожидалась открывающая скобка '(' после ключа атрибута", ch, i)
			} else {
				p.buf = append(p.buf, ch)
			}
		case READ_ATTR_VAL:
			if ch == ')' {
				p.curTask.Attributes = append(p.curTask.Attributes, domain.Attribute{
					Key:   p.attrKey,
					Value: string(p.buf),
				})
				p.buf = p.buf[:0]
				// p.attrKey = ""
				p.state = READ_ATTR_WAIT_SPACE
			} else if ch == '\n' {
				return nil, p.syntaxError("ожидалась закрывающая скобка ')' для значения атрибута", ch, i)
			} else {
				p.buf = append(p.buf, ch)
			}
		case READ_ATTR_WAIT_SPACE:
			if ch == ' ' || ch == '\t' {
			} else if ch == '#' {
				p.buf = p.buf[:0]
				p.state = READ_TASK_TAG
			} else if ch == '@' {
				p.buf = p.buf[:0]
				p.state = READ_ATTR_KEY
			} else if ch == '\n' || ch == '\r' {
				p.state = IN_TASKS
			} else {
				return nil, p.syntaxError("после атрибутов допустимы только другие атрибуты (@), теги (#) или перенос строки", ch, i)
			}
		case IN_DAILY_LOG:
			if ch == '#' {
				if len(p.buf) > 0 {
					p.doc.DailyLog = append(p.doc.DailyLog, string(p.buf))
					p.buf = p.buf[:0]
				}
				p.state = SEEN_H1
			} else if ch == '\n' {
				if len(p.buf) > 0 {
					p.doc.DailyLog = append(p.doc.DailyLog, string(p.buf))
					p.buf = p.buf[:0]
				}
			} else {
				p.buf = append(p.buf, ch)
			}
		}
	}
	if p.state == READ_TASK_TEXT && p.hasTask {
		p.curTask.Title = trimTrailingSpaces(string(p.buf))
	} else if p.state == READ_TASK_TAG && len(p.buf) > 0 {
		p.curTask.Tags = append(p.curTask.Tags, string(p.buf))
	} else if p.state == READ_NOTE_TEXT && len(p.buf) > 0 {
		p.curTask.Notes = append(p.curTask.Notes, trimTrailingSpaces(string(p.buf)))
	} else if p.state == IN_DAILY_LOG && len(p.buf) > 0 {
		p.doc.DailyLog = append(p.doc.DailyLog, trimTrailingSpaces(string(p.buf)))
	}

	p.saveCurrentTask()

	switch p.state {
	case IN_TASKS, IN_DAILY_LOG, WAIT_SECTION, READ_TASK_TEXT, READ_TASK_TAG, READ_ATTR_WAIT_SPACE, READ_NOTE_TEXT, START:
		return &p.doc, nil
	default:
		return nil, fmt.Errorf("неожиданный конец файла: незавершенная конструкция (state: %v)", p.state)
	}
}

func (p *Parser) syntaxError(msg string, got byte, pos int) error {
	p.state = STATE_ERROR
	return fmt.Errorf("ошибка на позиции %d: Syntax error: %s, got: %q", pos, msg, got)
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func (p *Parser) saveCurrentTask() {
	if p.hasTask {
		p.doc.Tasks = append(p.doc.Tasks, p.curTask)
		p.curTask = domain.Task{}
		p.hasTask = false
	}
}

func trimTrailingSpaces(s string) string {
	i := len(s) - 1
	for i >= 0 && (s[i] == ' ' || s[i] == '\t' || s[i] == '\r') {
		i--
	}
	return s[:i+1]
}
