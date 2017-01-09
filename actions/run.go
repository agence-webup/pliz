package actions

import (
	"fmt"
	"webup/pliz/domain"

	"github.com/fatih/color"
)

func RunTaskActionHandler(task domain.Task, args *[]string, prod bool) func() {
	return func() {

		// disable the execution check for standalone execution
		task.ExecutionCheck = nil

		// add the additionnal arguments if needed
		if args != nil {
			task.AdditionnalArgs = *args
		}

		if task.Execute(domain.TaskExecutionContext{Prod: prod}) {
			fmt.Printf("Task '%s' %s.\n", task.Name, color.GreenString("executed"))
		}
	}
}
