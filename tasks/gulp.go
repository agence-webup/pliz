package tasks

import "webup/pliz/domain"

// Create a task for running 'gulp'
func GulpTask(container string) domain.Task {
	task := domain.Task{Name: "gulp", Description: "Run 'gulp' in the build container"}

	// execute 'gulp' into the builder container
	task.Container = &container
	task.CommandArgs = []string{"./node_modules/.bin/gulp"}

	return task
}
