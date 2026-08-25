package core

import (
	"fmt"
)

func (doc *DayDocument) AddTask(newTask Task) error {
	if newTask.Title == "" {
		return fmt.Errorf("задача не может быть без названия")
	}
	doc.Tasks = append(doc.Tasks, newTask)
	return doc.Save()
}

func (doc *DayDocument) RemoveTask(taskIndex int) error {
	if taskIndex < 0 || taskIndex >= len(doc.Tasks) {
		return fmt.Errorf("неверный индекс задачи")
	}

	taskToRemove := doc.Tasks[taskIndex]

	doc.Tasks = append(doc.Tasks[:taskIndex], doc.Tasks[taskIndex+1:]...)

	if taskToRemove.Done && doc.Meta.TotalCompleted > 0 {
		doc.Meta.TotalCompleted--
	}

	return doc.Save()
}

func (doc *DayDocument) ToggleTaskStatus(taskIndex int) error {
	if taskIndex < 0 || taskIndex >= len(doc.Tasks) {
		return fmt.Errorf("индекс задачи %d вне диапазона", taskIndex)
	}

	task := &doc.Tasks[taskIndex]
	task.Done = !task.Done

	if task.Done {
		doc.Meta.TotalCompleted++
	} else {
		if doc.Meta.TotalCompleted > 0 {
			doc.Meta.TotalCompleted--
		}
	}

	return doc.Save()
}
