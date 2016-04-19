package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"webup/pliz/domain"
	"webup/pliz/tasks"

	"gopkg.in/yaml.v2"
)

const (
	defaultFilename = "pliz.yml"
)

var loadedConfig *domain.Config

func Check() error {

	if loadedConfig == nil {
		config, err := parseConfigFile()
		if err != nil {
			return err
		}
		loadedConfig = &config
	}

	if _, err := os.Stat("docker-compose.yml"); os.IsNotExist(err) {
		fmt.Println("Unable to find a Docker Compose file in the current directory")
		return err
	}

	return nil
}

func Get() domain.Config {
	return *loadedConfig
}

type parserConfig struct {
	Containers   map[string]string `yaml:"containers"`
	ConfigFiles  map[string]string `yaml:"config_files"`
	Tasks        []TaskSpec        `yaml:"tasks"`
	InstallTasks []domain.TaskID   `yaml:"install_tasks"`
	Checklist    []string          `yaml:"checklist"`
	Backup       BackupSpec        `yaml:"backup"`
}

func (parsed parserConfig) convertToConfig(config *domain.Config) error {
	// container config
	containerConfig := domain.ContainerConfig{
		Proxy:   "proxy",
		App:     "app",
		Builder: "srcbuild",
		Db:      "db",
	}
	if proxyContainerName, ok := parsed.Containers["proxy"]; ok {
		containerConfig.Proxy = proxyContainerName
	}
	if appContainerName, ok := parsed.Containers["app"]; ok {
		containerConfig.App = appContainerName
	}
	if builderContainerName, ok := parsed.Containers["builder"]; ok {
		containerConfig.Builder = builderContainerName
	}
	config.Containers = containerConfig

	// config files
	configFiles := []domain.ConfigFile{}
	for sample, target := range parsed.ConfigFiles {
		configFiles = append(configFiles, domain.ConfigFile{Sample: sample, Target: target})
	}
	config.ConfigFiles = configFiles

	// will be used to prepare the list of the available tasks (default & custom tasks)
	tasksByID := map[domain.TaskID]domain.Task{}
	for _, id := range domain.DefaultTaskNames() {
		task, err := tasks.CreateTaskWithName(id, *config)
		if err != nil {
			return err
		}

		tasksByID[id] = task
	}

	// tasks
	for i, taskSpec := range parsed.Tasks {

		id := domain.TaskID(taskSpec.Name)

		// check if the task is an overrided default task
		if task, ok := tasksByID[id]; ok {

			// overrided description (not required)
			if taskSpec.Description != "" {
				task.Description = taskSpec.Description
			}

			// check if a container is specified
			// if the value is 'none', the command will be run on the host
			if taskSpec.Container != "" {
				if taskSpec.Container != "none" {
					task.Container = &(parsed.Tasks[i].Container)
				} else {
					task.Container = nil
				}
			}

			// check if the command is overrided
			if len(taskSpec.CommandArgs) > 0 {
				task.CommandArgs = taskSpec.CommandArgs
			} else {
				return fmt.Errorf("Not enough args to execute the command of the task '%s'", taskSpec.Name)
			}

			tasksByID[id] = task
		} else {
			// it's a custom task

			// check if the custom task is valid
			if err := taskSpec.IsValidForCustomTask(); err != nil {
				return fmt.Errorf("Custom task error: %v", err)
			}

			task := domain.Task{Name: id, Description: taskSpec.Description}
			// check if the container is specified
			if taskSpec.Container != "none" {
				task.Container = &(parsed.Tasks[i].Container)
			}
			// command args
			task.CommandArgs = taskSpec.CommandArgs

			tasksByID[id] = task
		}
	}
	config.Tasks = tasksByID

	// install tasks
	for _, id := range parsed.InstallTasks {
		if _, ok := config.Tasks[id]; !ok {
			return fmt.Errorf("Install tasks: '%s' is not available", id)
		}
	}
	config.InstallTasks = parsed.InstallTasks

	// checklist
	config.Checklist = parsed.Checklist

	// backup
	backupConfig := domain.Backup{Files: parsed.Backup.Files, Databases: []domain.DatabaseBackupConfig{}}
	for i := range parsed.Backup.Databases {
		dbBackupConfig := domain.DatabaseBackupConfig{Container: parsed.Backup.Databases[i].Container, Type: parsed.Backup.Databases[i].Type}
		backupConfig.Databases = append(backupConfig.Databases, dbBackupConfig)
	}
	config.BackupConfig = backupConfig

	return nil
}

func parseConfigFile() (domain.Config, error) {

	config := domain.Config{}

	configFile, err := ioutil.ReadFile(defaultFilename)
	if err != nil {
		fmt.Println("Unable to find a config file 'pliz.yml' in the current directory")
		return config, err
	}

	var parsed parserConfig
	err = yaml.Unmarshal(configFile, &parsed)
	if err != nil {
		fmt.Println("Unable to parse the config file. Check 'pliz.yml' syntax.")
		fmt.Println(err)
		return config, err
	}

	err = parsed.convertToConfig(&config)
	if err != nil {
		fmt.Println(err)
		return config, err
	}

	return config, nil
}
