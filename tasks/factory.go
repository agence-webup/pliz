package tasks

import (
	"fmt"
	"webup/pliz/domain"
)

func CreateTaskWithName(name domain.TaskID, config domain.Config) (domain.Task, error) {

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
	case "db:update":
		return DbUpdateTask(config.Containers.App), nil
	case "git:plug-hook":
		return GitPlugHookTask(), nil
	}

	return domain.Task{}, fmt.Errorf("Unable to find the task '%s'\n", name)
}
