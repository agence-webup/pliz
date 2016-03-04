package config

type TaskSpec struct {
	Name     string            `yaml:"name"`
	Override *TaskOverrideSpec `yaml:"override"`
}

type TaskOverrideSpec struct {
	Container   *string
	CommandArgs *[]string `yaml:"command"`
}
