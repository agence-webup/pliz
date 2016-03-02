package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	defaultFilename = "pliz.yml"
)

type Config struct {
	ConfigFiles []ConfigFile `yaml:"config_files"`

	// backward compatibility
	Default DefaultActionsConfig `yaml:"default"`
	Custom  []CustomAction       `yaml:"custom"`
}

type ConfigFile struct {
	Sample string
	Target string
}

func GetConfig(config *Config) error {

	configFile, err := ioutil.ReadFile(defaultFilename)
	if err != nil {
		fmt.Println("Unable to find a config file 'pliz.yml' in the current directory")
		return err
	}

	// fmt.Println(string(configFile))
	err = yaml.Unmarshal(configFile, config)
	if err != nil {
		fmt.Println("Unable to parse the config file. Check the file's syntax.")
		fmt.Println(err)
		return err
	}

	return nil
}
