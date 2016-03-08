package domain

import "fmt"

type Task struct {
	Name           string
	Description    string
	Container      *string
	ExecutionCheck TaskExecutionCheck
	CommandArgs    CommandArgs
}

func (t Task) Execute() bool {
	if t.ExecutionCheck != nil && !t.ExecutionCheck.CanExecute() {
		// return errors.New(fmt.Sprintf("Task '%s' skipped.", t.Name))
		fmt.Printf("Task '%s' skipped.\n", t.Name)
		return false
	}

	var command Command
	if t.Container != nil {
		command = NewContainerCommand(*t.Container, t.CommandArgs)
	} else {
		command = NewCommand(t.CommandArgs)
	}
	command.Execute()

	if t.ExecutionCheck != nil {
		t.ExecutionCheck.PostExecute()
	}

	return true
}
