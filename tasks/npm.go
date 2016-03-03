package tasks

import "webup/pliz/helpers"

// Create a task for running 'npm install'
func NpmTask() Task {
	task := Task{Name: "npm", Description: "Run 'npm install' in the build container"}

	task.ExecutionCheck = &ModificationDateTaskExecutionCheck{UpdatedFile: "package.json", CompareTo: "node_modules"}
	task.Command = helpers.NewContainerCommand("srcbuild", []string{"npm", "install"})

	return task
}
