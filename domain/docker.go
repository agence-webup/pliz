package domain

type DockerContainerEnv map[string]string

type DockerContainerConfig struct {
	Image string
	Env   DockerContainerEnv
}
