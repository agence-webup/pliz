package tasks

import "webup/pliz/helpers"

// Create a task for running 'npm install'
func BowerTask() Task {
	task := Task{Name: "bower", Description: "Run 'bower install' in the build container"}

	task.ExecutionCheck = &ModificationDateTaskExecutionCheck{UpdatedFile: "bower.json", CompareTo: "public/bower"}
	task.Command = helpers.NewContainerCommand("srcbuild", []string{"bower", "install", "--allow-root"})

	return task
}
