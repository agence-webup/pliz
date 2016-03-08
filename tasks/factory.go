package tasks

import (
	"errors"
	"fmt"
	"webup/pliz/domain"
)

func CreateTaskWithName(name string, config domain.Config) (domain.Task, error) {

	// default tasks
	switch name {
	case "npm":
		return NpmTask(config.Containers.Builder), nil
	case "bower":
		return BowerTask(config.Containers.Builder), nil
	case "composer":
		return ComposerTask(config.Containers.App), nil
	case "gulp":
		return GulpTask(config.Containers.Builder), nil
	case "db-update":
		return DbUpdateTask(config.Containers.App), nil
	}

	// custom tasks
	for _, task := range config.CustomTasks {
		if name == task.Name {
			return task, nil
		}
	}

	return domain.Task{}, errors.New(fmt.Sprintf("Unable to find the task '%s'\n", name))
}
