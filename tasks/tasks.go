package tasks

import (
	"errors"
	"fmt"
	"webup/pliz/domain"
)

func CreateTaskWithName(name string, config domain.Config) (domain.Task, error) {
	switch name {
	case "npm":
		return NpmTask(config.Containers.Builder), nil
	case "bower":
		return BowerTask(config.Containers.Builder), nil
	}

	return domain.Task{}, errors.New(fmt.Sprintf("Unable to find the task '%s'\n", name))
}
