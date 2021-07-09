package tasks

import "webup/pliz/domain"

// Create a task for running 'gulp'
func GitPlugHookTask() domain.Task {
	task := domain.Task{Name: "git:plug-hook", Description: "Run 'pliz git:plug-hook'"}

	task.CommandArgs = []string{"pliz", "git:plug-hook"}

	return task
}
