package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"webup/pliz/domain"
)

type containerParsedConfig struct {
	Env   []string
	Image string
}

func GetContainerID(container string, ctx domain.ExecutionContext) string {
	cmd := domain.NewComposeCommand([]string{"ps", "-q", container}, ctx.IsProd())
	containerID, err := cmd.GetResult()
	if err != nil {
		fmt.Println("Unable to get the 'db' container id")
	}

	return containerID
}

func GetContainerConfig(containerID string, ctx domain.ExecutionContext) domain.DockerContainerConfig {
	cmd := domain.NewCommand([]string{"docker", "inspect", "--format", "{{json .Config}}", containerID})
	configJson, err := cmd.GetResult()
	if err != nil {
		fmt.Println("Unable to get the config of the 'db' container")
	}

	// parse the json
	var config containerParsedConfig
	json.NewDecoder(strings.NewReader(configJson)).Decode(&config)

	// parse env variables of the container
	env := domain.DockerContainerEnv{}
	for _, data := range config.Env {
		items := strings.SplitN(data, "=", 2)
		env[items[0]] = items[1]
	}

	return domain.DockerContainerConfig{
		Image: config.Image,
		Env:   env,
	}
}

func GetExposedPorts(containerID string, ctx domain.ExecutionContext) []string {
	cmd := domain.NewCommand([]string{"docker", "inspect", "--format", "{{json .NetworkSettings.Ports}}", containerID})
	configJson, err := cmd.GetResult()
	if err != nil {
		fmt.Println("Unable to get the network settings of the container")
	}

	// parse the json
	var networkSettings map[string][]map[string]string
	json.NewDecoder(strings.NewReader(configJson)).Decode(&networkSettings)

	ports := []string{}
	for _, portSettings := range networkSettings {
		for _, settings := range portSettings {
			ports = append(ports, settings["HostPort"])
		}
	}

	return ports
}
