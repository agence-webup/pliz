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
	EnabledTasks []TaskSpec        `yaml:"enabled_tasks"`
	Checklist    []string          `yaml:"checklist"`
	CustomTasks  []CustomTaskSpec  `yaml:"custom_tasks"`
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

	// custom tasks
	// NOTE: must be handled before the enabled tasks because the custom tasks can be specified inside the enabled tasks
	customTasks := []domain.Task{}
	for i, taskSpec := range parsed.CustomTasks {

		// check if the custom task is valid
		if taskSpec.IsValid() {
			task := domain.Task{Name: taskSpec.Name, Description: taskSpec.Description}
			// check if the container is specified
			if taskSpec.Container != "none" {
				task.Container = &(parsed.CustomTasks[i].Container)
			}
			// command args
			task.CommandArgs = taskSpec.CommandArgs
			customTasks = append(customTasks, task)
		}
	}
	config.CustomTasks = customTasks

	// enabled tasks
	enabledTasks := []domain.Task{}
	for _, taskSpec := range parsed.EnabledTasks {

		task, err := tasks.CreateTaskWithName(taskSpec.Name, *config)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// check for override
		if taskSpec.Override != nil {

			// check if a container is specified
			// if the value is 'none', the command will be run on the host
			if taskSpec.Override.Container != nil {
				if *taskSpec.Override.Container != "none" {
					task.Container = taskSpec.Override.Container
				} else {
					task.Container = nil
				}
			}

			// check if the command is overrided
			if taskSpec.Override.CommandArgs != nil && len(*taskSpec.Override.CommandArgs) > 0 {
				task.CommandArgs = *taskSpec.Override.CommandArgs
			} else {
				fmt.Println("Not enough args to execute the command")
			}
		}

		enabledTasks = append(enabledTasks, task)
	}
	config.EnabledTasks = enabledTasks

	// checklist
	config.Checklist = parsed.Checklist

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

	parsed.convertToConfig(&config)

	return config, nil
}
