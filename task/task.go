package task

import (
	"errors"
	"fmt"
	"qcron/libs"
)

type Task struct {
	TimeExpress string
	TaskID string
	Command func(args ...interface{}) error
	Arguments []interface{}
	Description string
	TaskTime *libs.TaskTime
}

func NewTask(timeExpress, taskId string,
	command func(args ...interface{}) error, args []interface{},
	description string) (*Task, error) {
	task := &Task{
		TimeExpress: timeExpress,
		TaskID: taskId,
		Command: command,
		Arguments: args,
		Description: description,
	}

	tTime, err := libs.TimeParse(timeExpress)
	if err != nil {
		return nil, err
	}
	task.TaskTime = tTime
	return task, nil
}

func (t *Task) Run() error {
	err := t.Command(t.Arguments...)
	t.TaskTime.ComputeNextExecTime()
	return err
}

type TManager struct {
	Tasks []*Task
}

func NewTaskManager() *TManager {
	return &TManager{
		Tasks: make([]*Task, 0, 10),
	}
}

func (tm *TManager) Register(task *Task) error {
	for _, v := range tm.Tasks {
		if v.TaskID == task.TaskID {
			return errors.New(fmt.Sprintf("duplicate task, id: %s", task.TaskID))
		}
	}
	tm.Tasks = append(tm.Tasks, task)
	return nil
}

func (tm *TManager) UnRegister(taskId string) error {
	var index = -1
	for i, v := range tm.Tasks {
		if v.TaskID == taskId {
			index = i
			break
		}
	}

	if index == -1 {
		return errors.New(fmt.Sprintf("task id not found, id: %s", taskId))
	}

	// delete
	newTasks := make([]*Task, len(tm.Tasks) - 1)
	copy(newTasks[:index], tm.Tasks[:index])
	copy(newTasks[index:], tm.Tasks[index+1:])
	tm.Tasks = newTasks
	return nil
}

// 单例
var Manager = NewTaskManager()
