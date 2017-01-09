package domain

import (
	"fmt"
	"strings"
)

type TaskID string

type Task struct {
	Name            TaskID
	Description     string
	Container       *string
	ExecutionCheck  TaskExecutionCheck
	CommandArgs     CommandArgs
	AdditionnalArgs CommandArgs
}

func DefaultTaskNames() []TaskID {
	return []TaskID{
		"npm",
		"bower",
		"composer",
		"gulp",
		"db:update",
	}
}

func (t Task) Execute(context TaskExecutionContext) bool {
	if t.ExecutionCheck != nil && !t.ExecutionCheck.CanExecute() {
		// return errors.New(fmt.Sprintf("Task '%s' skipped.", t.Name))
		fmt.Printf("Task '%s' skipped.\n", t.Name)
		return false
	}

	args := append(t.CommandArgs, t.AdditionnalArgs...)

	var command Command
	if t.Container != nil {
		command = NewContainerCommand(*t.Container, args, []string{}, context.Prod)
	} else {
		command = NewCommand(args)
	}
	command.Execute()

	if t.ExecutionCheck != nil {
		t.ExecutionCheck.PostExecute()
	}

	return true
}

func (t Task) String() string {
	return fmt.Sprintf("%s => container:%v | %s", t.Name, *t.Container, strings.Join(t.CommandArgs, " "))
}

type TaskExecutionContext struct {
	Prod bool
}
