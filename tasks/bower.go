package tasks

import (
	"webup/pliz/config"
	"webup/pliz/helpers"
)

// Create a task for running 'npm install'
func BowerTask() Task {
	task := Task{Name: "bower", Description: "Run 'bower install' in the build container"}

	// check if 'bower.json' has been updated since last install into 'public/bower'
	task.ExecutionCheck = &ModificationDateTaskExecutionCheck{UpdatedFile: "bower.json", CompareTo: "public/bower"}

	// execute 'bower install' into the builder container
	builderContainer := config.Get().Containers.Builder
	task.Command = helpers.NewContainerCommand(builderContainer, []string{"bower", "install", "--allow-root"})

	return task
}
