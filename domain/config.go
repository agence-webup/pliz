package domain

type Config struct {
	Containers  ContainerConfig
	ConfigFiles []ConfigFile
	Tasks       map[TaskID]Task
	Checklist   []string

	InstallTasks []TaskID // list of tasks that will be executed during install

	BackupConfig Backup
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

type Backup struct {
	Files     []string
	Databases []DatabaseBackupConfig
}

type DatabaseBackupConfig struct {
	Container string
	Type      string
}
