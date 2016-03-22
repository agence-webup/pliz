package actions

import (
	"fmt"
	"webup/pliz/domain"
)

func RunTaskActionHandler(task domain.Task, prod bool) func() {
	return func() {

		// disable the execution check for standalone execution
		task.ExecutionCheck = nil

		if task.Execute(domain.TaskExecutionContext{Prod: prod}) {
			fmt.Printf("Task '%s' executed.\n", task.Name)
		}
	}
}
