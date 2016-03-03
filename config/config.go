package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	defaultFilename = "pliz.yml"
)

var loadedConfig *Config

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
	EnabledTasks []string

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
}

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

func Get() Config {
	return *loadedConfig
}

type parserConfig struct {
	Containers   map[string]string `yaml:"containers"`
	ConfigFiles  map[string]string `yaml:"config_files"`
	EnabledTasks []string          `yaml:"enabled_tasks"`
}

func (parsed parserConfig) convertToConfig(config *Config) error {
	// container config
	containerConfig := ContainerConfig{
		Proxy:   "proxy",
		App:     "app",
		Builder: "srcbuild",
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
	configFiles := []ConfigFile{}
	for sample, target := range parsed.ConfigFiles {
		configFiles = append(configFiles, ConfigFile{Sample: sample, Target: target})
	}
	config.ConfigFiles = configFiles

	// enabled tasks
	config.EnabledTasks = parsed.EnabledTasks

	return nil
}

func parseConfigFile() (Config, error) {

	config := Config{}

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
