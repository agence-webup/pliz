package tasks

import "webup/pliz/domain"

// Create a task for running 'npm install'
func NpmTask(container string) domain.Task {
	task := domain.Task{Name: "npm", Description: "Run 'npm install' in the build container"}

	// check if 'package.json' has been updated since last install into 'node_modules'
	task.ExecutionCheck = &domain.ModificationDateTaskExecutionCheck{UpdatedFile: "package.json", CompareTo: "node_modules"}

	// execute 'npm install' into the builder container
	task.Container = &container
	task.CommandArgs = []string{"npm", "install"}

	return task
}
