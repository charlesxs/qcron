package task

import (
	"errors"
	"fmt"
	"github.com/charlesxs/qcron/libs"
	"log"
	"time"
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

	tTime, err := libs.TimeParse(timeExpress, time.Now())
	if err != nil {
		return nil, err
	}
	task.TaskTime = tTime
	return task, nil
}

func (t *Task) Run() error {
	t.TaskTime.ComputeNextExecTime()
	err := t.Command(t.Arguments...)
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

func UpdateTasks()  {
	// 更新所有task 的 next execute time
	for i := range Manager.Tasks {
		times, ok := libs.InfoCache.Get(Manager.Tasks[i].TaskID)
		if !ok {
			continue
		}

		maxTime := GetMax(times)
		newTaskTime, err := libs.TimeParse(Manager.Tasks[i].TimeExpress, maxTime)
		if err != nil {
			log.Printf("update task: %s error, %s", Manager.Tasks[i].TaskID, err)
		}
		Manager.Tasks[i].TaskTime = newTaskTime

	}
	// clean old cache
	libs.InfoCache.CleanCache()
}

func GetMax(times []time.Time) time.Time {
	var (
		max int64
		index int
	)

	for i, v := range times {
		timestamp := v.Unix()
		if timestamp > max {
			max = timestamp
			index = i
		}
	}
	return times[index]
}


// 单例
var Manager = NewTaskManager()
