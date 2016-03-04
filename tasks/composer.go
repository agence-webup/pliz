package tasks

import "webup/pliz/domain"

// Create a task for running 'npm install'
func ComposerTask(container string) domain.Task {
	task := domain.Task{Name: "composer", Description: "Run 'composer install' in the app container"}

	// check if 'composer.json' has been updated since last install into 'vendor'
	task.ExecutionCheck = &domain.ModificationDateTaskExecutionCheck{UpdatedFile: "composer.json", CompareTo: "vendor"}

	// execute 'composer install' into the builder container
	task.Container = &container
	task.CommandArgs = []string{"composer", "install"}

	return task
}
