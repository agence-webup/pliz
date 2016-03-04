package tasks

import "webup/pliz/domain"

// Create a task for running 'npm install'
func DbUpdateTask(container string) domain.Task {
	task := domain.Task{Name: "db-update", Description: "Run the migrations to update the DB"}

	// execute 'php artisan migrate' into the app container
	task.Container = &container
	task.CommandArgs = []string{"php", "artisan", "migrate"}

	return task
}
