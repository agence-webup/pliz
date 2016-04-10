package config

import "errors"

type TaskSpec struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Container   string   `yaml:"container"`
	CommandArgs []string `yaml:"command"`
}

func (task TaskSpec) IsValidForCustomTask() error {
	if task.Name == "" {
		return errors.New("'name' is required")
	}
	if task.Container == "" {
		return errors.New("'container' is required")
	}
	if len(task.CommandArgs) == 0 {
		return errors.New("'command' is required")
	}

	return nil
}

type BackupSpec struct {
	Files     []string             `yaml:"files"`     // list of the files/directories to backup
	Databases []DatabaseBackupSpec `yaml:"databases"` // list of the db to backup
}

type DatabaseBackupSpec struct {
	Container string `yaml:"container"`
	Type      string `yaml:"type"`
}
