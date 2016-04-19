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
