package tasks

import "webup/pliz/domain"

// Create a task for running 'npm install'
func BowerTask(container string) domain.Task {
	task := domain.Task{Name: "bower", Description: "Run 'bower install' in the build container"}

	// check if 'bower.json' has been updated since last install into 'public/bower'
	task.ExecutionCheck = &domain.ModificationDateTaskExecutionCheck{UpdatedFile: "bower.json", CompareTo: "public/bower"}

	// execute 'bower install' into the builder container
	task.Container = &container
	task.CommandArgs = []string{"bower", "install", "--allow-root"}

	return task
}
