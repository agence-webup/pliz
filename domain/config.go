package domain

type Config struct {
	Containers   ContainerConfig
	ConfigFiles  []ConfigFile
	EnabledTasks []Task
	Checklist    []string
	CustomTasks  []Task
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
	return []string{c.Proxy, c.App, c.Builder, c.Db}
}
