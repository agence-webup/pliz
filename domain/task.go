package domain

import (
	"fmt"
	"strings"
)

type Task struct {
	Name           string
	Description    string
	Container      *string
	ExecutionCheck TaskExecutionCheck
	CommandArgs    CommandArgs
}

func (t Task) Execute(context TaskExecutionContext) bool {
	if t.ExecutionCheck != nil && !t.ExecutionCheck.CanExecute() {
		// return errors.New(fmt.Sprintf("Task '%s' skipped.", t.Name))
		fmt.Printf("Task '%s' skipped.\n", t.Name)
		return false
	}

	var command Command
	if t.Container != nil {
		command = NewContainerCommand(*t.Container, t.CommandArgs, []string{}, context.Prod)
	} else {
		command = NewCommand(t.CommandArgs)
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
