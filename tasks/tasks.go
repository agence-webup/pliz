package tasks

import (
	"errors"
	"fmt"
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

func CreateTaskWithName(name string) (Task, error) {
	switch name {
	case "npm":
		return NpmTask(), nil
	case "bower":
		return BowerTask(), nil
	}

	return Task{}, errors.New(fmt.Sprintf("Unable to find the task '%s'\n", name))
}
