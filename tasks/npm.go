package tasks

import (
	"webup/pliz/config"
	"webup/pliz/helpers"
)

// Create a task for running 'npm install'
func NpmTask() Task {
	task := Task{Name: "npm", Description: "Run 'npm install' in the build container"}

	// check if 'package.json' has been updated since last install into 'node_modules'
	task.ExecutionCheck = &ModificationDateTaskExecutionCheck{UpdatedFile: "package.json", CompareTo: "node_modules"}

	// execute 'npm install' into the builder container
	builderContainer := config.Get().Containers.Builder
	task.Command = helpers.NewContainerCommand(builderContainer, []string{"npm", "install"})

	return task
}
