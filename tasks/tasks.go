package tasks

import (
	"fmt"
	"os"
	"time"
	"webup/pliz/helpers"
)

type Task struct {
	Name           string
	Description    string
	ExecutionCheck TaskExecutionCheck
	Command        helpers.Command
}

func (t Task) Execute() bool {
	if t.ExecutionCheck != nil && !t.ExecutionCheck.CanExecute() {
		// return errors.New(fmt.Sprintf("Task '%s' skipped.", t.Name))
		fmt.Printf("Task '%s' skipped.\n", t.Name)
		return false
	}

	t.Command.Execute()

	if t.ExecutionCheck != nil {
		t.ExecutionCheck.PostExecute()
	}

	return true
}

// type Task interface {
// 	GetName() string
// 	GetDescription() string
// 	GetExecutionCheck() *TaskExecutionCheck
//
// 	Execute() error
// }

type TaskExecutionCheck interface {
	CanExecute() bool
	PostExecute()
}

type ModificationDateTaskExecutionCheck struct {
	UpdatedFile string
	CompareTo   string
}

// Check if modification time of the file 'UpdatedFile' is later than the file 'CompareTo'
func (chk ModificationDateTaskExecutionCheck) CanExecute() bool {
	updatedFileStat, err := os.Stat(chk.UpdatedFile)
	if err != nil {
		fmt.Println(err)
		return false
	}
	compareToStat, err := os.Stat(chk.CompareTo)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if updatedFileStat.ModTime().Before(compareToStat.ModTime()) {
		return false
	}

	return true
}

func (chk ModificationDateTaskExecutionCheck) PostExecute() {
	currentTime := time.Now().Local()
	os.Chtimes(chk.CompareTo, currentTime, currentTime)
}
