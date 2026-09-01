package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"context"
	"encoding/json"
	"jrn/internal/parser"
	"jrn/pkg/domain"
	"jrn/pkg/engine"
	"jrn/pkg/storage"
	"sync"
	"time"
	"unsafe"
)

var (
	defaultEngine *engine.Engine
	mu            sync.Mutex
)

//export JRN_Init
func JRN_Init(path *C.char) int {
	mu.Lock()
	defer mu.Unlock()

	dir := C.GoString(path)
	store, err := storage.New(dir)
	if err != nil {
		return -1
	}
	defaultEngine = engine.New(store)
	return 0
}

//export JRN_OpenTodayJSON
func JRN_OpenTodayJSON() *C.char {
	mu.Lock()
	defer mu.Unlock()

	if defaultEngine == nil {
		return C.CString(`{"error":"engine not initialized"}`)
	}

	doc, err := defaultEngine.OpenToday(context.Background())
	if err != nil {
		return C.CString(`{"error":"` + err.Error() + `"}`)
	}

	return C.CString(string(parser.SerializeJson(doc)))
}

//export JRN_OpenDayJSON
func JRN_OpenDayJSON(dateStr *C.char) *C.char {
	mu.Lock()
	defer mu.Unlock()

	if defaultEngine == nil {
		return C.CString(`{"error":"engine not initialized"}`)
	}

	dStr := C.GoString(dateStr)
	t, err := time.Parse("2006-01-02", dStr)
	if err != nil {
		return C.CString(`{"error":"invalid date format, expected YYYY-MM-DD"}`)
	}

	doc, err := defaultEngine.OpenDay(context.Background(), t)
	if err != nil {
		return C.CString(`{"error":"` + err.Error() + `"}`)
	}

	return C.CString(string(parser.SerializeJson(doc)))
}

//export JRN_AddTask
func JRN_AddTask(dateStr *C.char, taskJSON *C.char) int {
	mu.Lock()
	defer mu.Unlock()

	if defaultEngine == nil {
		return -1
	}

	t, err := time.Parse("2006-01-02", C.GoString(dateStr))
	if err != nil {
		return -2
	}

	var task domain.Task
	if err := json.Unmarshal([]byte(C.GoString(taskJSON)), &task); err != nil {
		return -3
	}

	doc, err := defaultEngine.OpenDay(context.Background(), t)
	if err != nil {
		return -4
	}

	if err := doc.AddTask(task); err != nil {
		return -5
	}

	if err := defaultEngine.SaveDay(context.Background(), doc); err != nil {
		return -6
	}

	return 0
}

//export JRN_ToggleTask
func JRN_ToggleTask(dateStr *C.char, index int) int {
	mu.Lock()
	defer mu.Unlock()

	if defaultEngine == nil {
		return -1
	}

	t, err := time.Parse("2006-01-02", C.GoString(dateStr))
	if err != nil {
		return -2
	}

	doc, err := defaultEngine.OpenDay(context.Background(), t)
	if err != nil {
		return -3
	}

	if err := doc.ToggleTask(index); err != nil {
		return -4
	}

	if err := defaultEngine.SaveDay(context.Background(), doc); err != nil {
		return -5
	}

	return 0
}

//export JRN_DeleteTask
func JRN_DeleteTask(dateStr *C.char, index int) int {
	mu.Lock()
	defer mu.Unlock()

	if defaultEngine == nil {
		return -1
	}

	t, err := time.Parse("2006-01-02", C.GoString(dateStr))
	if err != nil {
		return -2
	}

	doc, err := defaultEngine.OpenDay(context.Background(), t)
	if err != nil {
		return -3
	}

	if err := doc.DeleteTask(index); err != nil {
		return -4
	}

	if err := defaultEngine.SaveDay(context.Background(), doc); err != nil {
		return -5
	}

	return 0
}

//export JRN_UpdateTaskTitle
func JRN_UpdateTaskTitle(dateStr *C.char, index int, title *C.char) int {
	mu.Lock()
	defer mu.Unlock()

	if defaultEngine == nil {
		return -1
	}

	t, err := time.Parse("2006-01-02", C.GoString(dateStr))
	if err != nil {
		return -2
	}

	doc, err := defaultEngine.OpenDay(context.Background(), t)
	if err != nil {
		return -3
	}

	if err := doc.UpdateTaskTitle(index, C.GoString(title)); err != nil {
		return -4
	}

	if err := defaultEngine.SaveDay(context.Background(), doc); err != nil {
		return -5
	}

	return 0
}

//export JRN_ReorderTasks
func JRN_ReorderTasks(dateStr *C.char, fromIndex int, toIndex int) int {
	mu.Lock()
	defer mu.Unlock()

	if defaultEngine == nil {
		return -1
	}

	t, err := time.Parse("2006-01-02", C.GoString(dateStr))
	if err != nil {
		return -2
	}

	doc, err := defaultEngine.OpenDay(context.Background(), t)
	if err != nil {
		return -3
	}

	if err := doc.ReorderTasks(fromIndex, toIndex); err != nil {
		return -4
	}

	if err := defaultEngine.SaveDay(context.Background(), doc); err != nil {
		return -5
	}

	return 0
}

//export JRN_AddDailyLog
func JRN_AddDailyLog(dateStr *C.char, text *C.char) int {
	mu.Lock()
	defer mu.Unlock()

	if defaultEngine == nil {
		return -1
	}

	t, err := time.Parse("2006-01-02", C.GoString(dateStr))
	if err != nil {
		return -2
	}

	doc, err := defaultEngine.OpenDay(context.Background(), t)
	if err != nil {
		return -3
	}

	doc.AddDailyLog(C.GoString(text))

	if err := defaultEngine.SaveDay(context.Background(), doc); err != nil {
		return -4
	}

	return 0
}

//export JRN_DeleteDailyLog
func JRN_DeleteDailyLog(dateStr *C.char, index int) int {
	mu.Lock()
	defer mu.Unlock()

	if defaultEngine == nil {
		return -1
	}

	t, err := time.Parse("2006-01-02", C.GoString(dateStr))
	if err != nil {
		return -2
	}

	doc, err := defaultEngine.OpenDay(context.Background(), t)
	if err != nil {
		return -3
	}

	if err := doc.DeleteDailyLog(index); err != nil {
		return -4
	}

	if err := defaultEngine.SaveDay(context.Background(), doc); err != nil {
		return -5
	}

	return 0
}

//export JRN_SearchJSON
func JRN_SearchJSON(query *C.char) *C.char {
	mu.Lock()
	defer mu.Unlock()

	if defaultEngine == nil {
		return C.CString(`{"error":"engine not initialized"}`)
	}

	results, err := defaultEngine.Search(context.Background(), C.GoString(query))
	if err != nil {
		return C.CString(`{"error":"` + err.Error() + `"}`)
	}

	data, err := json.Marshal(results)
	if err != nil {
		return C.CString(`{"error":"` + err.Error() + `"}`)
	}

	return C.CString(string(data))
}

//export JRN_GetTasksByTagJSON
func JRN_GetTasksByTagJSON(tag *C.char) *C.char {
	mu.Lock()
	defer mu.Unlock()

	if defaultEngine == nil {
		return C.CString(`{"error":"engine not initialized"}`)
	}

	results, err := defaultEngine.GetTasksByTag(context.Background(), C.GoString(tag))
	if err != nil {
		return C.CString(`{"error":"` + err.Error() + `"}`)
	}

	data, err := json.Marshal(results)
	if err != nil {
		return C.CString(`{"error":"` + err.Error() + `"}`)
	}

	return C.CString(string(data))
}

//export JRN_GetOverdueTasksJSON
func JRN_GetOverdueTasksJSON(todayStr *C.char) *C.char {
	mu.Lock()
	defer mu.Unlock()

	if defaultEngine == nil {
		return C.CString(`{"error":"engine not initialized"}`)
	}

	today, err := time.Parse("2006-01-02", C.GoString(todayStr))
	if err != nil {
		return C.CString(`{"error":"invalid date format, expected YYYY-MM-DD"}`)
	}

	results, err := defaultEngine.GetOverdueTasks(context.Background(), today)
	if err != nil {
		return C.CString(`{"error":"` + err.Error() + `"}`)
	}

	data, err := json.Marshal(results)
	if err != nil {
		return C.CString(`{"error":"` + err.Error() + `"}`)
	}

	return C.CString(string(data))
}

//export JRN_ListAvailableDaysJSON
func JRN_ListAvailableDaysJSON() *C.char {
	mu.Lock()
	defer mu.Unlock()

	if defaultEngine == nil {
		return C.CString(`{"error":"engine not initialized"}`)
	}

	dates, err := defaultEngine.ListAvailableDays(context.Background())
	if err != nil {
		return C.CString(`{"error":"` + err.Error() + `"}`)
	}

	strDates := make([]string, 0, len(dates))
	for _, d := range dates {
		strDates = append(strDates, d.Format("2006-01-02"))
	}

	data, err := json.Marshal(strDates)
	if err != nil {
		return C.CString(`{"error":"` + err.Error() + `"}`)
	}

	return C.CString(string(data))
}

//export JRN_FreeString
func JRN_FreeString(str *C.char) {
	C.free(unsafe.Pointer(str))
}

func main() {}
