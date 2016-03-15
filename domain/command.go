package domain

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CommandArgs []string

type Command struct {
	Name string
	Args []string
}

func (c Command) String() string {
	return fmt.Sprintf("%s %s", c.Name, strings.Join(c.Args, " "))
}

func (c Command) Execute() {
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Printf("Executing: %s\n", c)

	cmd.Run()
}

func NewCommand(list []string) Command {
	var name string
	var args []string

	if len(list) > 1 {
		name = list[0]
		args = list[1:]
	} else {
		name = list[0]
		args = []string{}
	}

	return Command{Name: name, Args: args}
}

func NewContainerCommand(container string, list []string, prod bool) Command {
	name := "docker-compose"

	var args []string
	if prod {
		args = []string{"-f", "docker-compose.yml", "-f", "docker-compose.prod.yml", "run", "--rm", container}
	} else {
		args = []string{"run", "--rm", container}
	}

	args = append(args, list...)

	return Command{Name: name, Args: args}
}
