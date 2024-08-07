package domain

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

type CommandArgs []string

type Command struct {
	Name    string
	Args    []string
	Verbose bool
}

func (c Command) String() string {
	return fmt.Sprintf("%s %s", c.Name, strings.Join(c.Args, " "))
}

func (c Command) Execute() {
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if c.Verbose {
		fmt.Printf("%s %s\n", color.MagentaString("Executing:"), c)
	}

	cmd.Run()
}

func (c Command) GetRawExecCommand() *exec.Cmd {
	return exec.Command(c.Name, c.Args...)
}

func (c Command) ExecuteWithStdin(reader io.Reader) {
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = reader

	if c.Verbose {
		fmt.Printf("%s %s\n", color.MagentaString("Executing:"), c)
	}

	cmd.Run()
}

func (c Command) GetResult() (string, error) {
	cmd := exec.Command(c.Name, c.Args...)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	output := strings.TrimSpace(string(out))

	return output, nil
}

func (c Command) WriteResultToFile(file *os.File) error {
	cmd := exec.Command(c.Name, c.Args...)

	cmd.Stdout = file
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	if c.Verbose {
		fmt.Printf("Executing: %s\n", c)
	}
	fmt.Printf("Writing to file: %s\n", file.Name())

	return nil
}

func NewCommand(list []string, verbose bool) Command {
	var name string
	var args []string

	if len(list) > 1 {
		name = list[0]
		args = list[1:]
	} else {
		name = list[0]
		args = []string{}
	}

	return Command{Name: name, Args: args, Verbose: verbose}
}

func NewComposeCommand(list []string, prod bool) Command {
	name := "docker"

	isProd := prod

	if isProd {
		if _, err := os.Stat("docker-compose.prod.yml"); os.IsNotExist(err) {
			fmt.Printf("\n%s: The file 'docker-compose.prod.yml' does not exist.\n", color.YellowString("Warning"))
			isProd = false
		}
	}

	args := []string{"compose"}
	if isProd {
		args = []string{"compose", "-f", "docker-compose.yml", "-f", "docker-compose.prod.yml"}
	}

	args = append(args, list...)

	return Command{Name: name, Args: args}
}

func NewContainerCommand(container string, list []string, options []string, prod bool) Command {
	args := []string{"run", "--rm"}

	// append the options
	args = append(args, options...)
	// the container
	args = append(args, container)
	// and the command args
	args = append(args, list...)

	return NewComposeCommand(args, prod)
}
