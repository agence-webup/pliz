package domain

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
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

	fmt.Printf("%s %s\n", color.MagentaString("Executing:"), c)

	cmd.Run()
}

func (c Command) GetRawExecCommand() *exec.Cmd {
	return exec.Command(c.Name, c.Args...)
}

func (c Command) ExecuteWithStdin(reader io.Reader) {
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	stdinReader := bufio.NewReader(reader)
	stdinWriter, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println(err)
	}

	_, err = stdinReader.WriteTo(stdinWriter)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%s %s\n", color.MagentaString("Executing:"), c)
	cmd.Start()

	// close writer to indicate that stdin is finished (avoiding hanging of the exec cmd)
	stdinWriter.Close()

	cmd.Wait()
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
	fmt.Printf("Executing: %s\n", c)
	fmt.Printf("Writing to file: %s\n", file.Name())

	return nil
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

func NewComposeCommand(list []string, prod bool) Command {
	name := "docker-compose"

	args := []string{}
	if prod {
		args = []string{"-f", "docker-compose.yml", "-f", "docker-compose.prod.yml"}
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
