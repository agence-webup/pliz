package domain

type DefaultActionsConfig struct {
	Install    Action `yaml:"install"`
	SrcPrepare Action `yaml:"src-prepare"`
}

type Action struct {
	Commands CommandList
}

type CommandList [][]string

type Config struct {
	Containers   ContainerConfig
	ConfigFiles  []ConfigFile
	EnabledTasks []Task

	// backward compatibility
	Default DefaultActionsConfig `yaml:"default"`
	Custom  []CustomAction       `yaml:"custom"`
}
type CustomAction struct {
	Name   string
	Action `yaml:",inline"`
}

type ConfigFile struct {
	Sample string
	Target string
}

type ContainerConfig struct {
	Proxy   string
	App     string
	Builder string
	Db      string
}

func (c ContainerConfig) All() []string {
	return []string{c.Proxy, c.App, c.Builder}
}
