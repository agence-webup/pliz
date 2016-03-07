package config

type TaskSpec struct {
	Name     string            `yaml:"name"`
	Override *TaskOverrideSpec `yaml:"override"`
}

type TaskOverrideSpec struct {
	Container   *string   `yaml:"container"`
	CommandArgs *[]string `yaml:"command"`
}

type CustomTaskSpec struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Container   string   `yaml:"container"`
	CommandArgs []string `yaml:"command"`
}

func (task CustomTaskSpec) IsValid() bool {
	if task.Name == "" || task.Container == "" || len(task.CommandArgs) == 0 {
		return false
	}
	return true
}
