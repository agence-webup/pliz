package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/jawher/mow.cli"
	"gopkg.in/yaml.v2"
)

type CommandList [][]string

type DefaultActionsConfig struct {
	Install    Action `yaml:"install"`
	SrcPrepare Action `yaml:"src-prepare"`
}

type Action struct {
	Commands CommandList
}
type CustomAction struct {
	Name   string
	Action `yaml:",inline"`
}

type Config struct {
	Default DefaultActionsConfig `yaml:"default"`
	Custom  []CustomAction       `yaml:"custom"`
}

type Command struct {
	Name string
	Args []string
}

func NewCommand(list []string) Command {
	name := list[0]
	args := list[1:]

	return Command{Name: name, Args: args}
}

func (c Command) Execute() {
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

func main() {
	app := cli.App("pliz", "Manage projects building")

	var config Config

	app.Before = func() {
		configFile, err := ioutil.ReadFile("pliz.yml")
		if err != nil {
			fmt.Println(err)
			return
		}

		// fmt.Println(string(configFile))
		err = yaml.Unmarshal(configFile, &config)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	app.Command("ps", "List running containers", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			command := Command{Name: "docker", Args: []string{"ps"}}
			command.Execute()
		}
	})

	app.Command("config", "...", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			fmt.Println(config)
		}
	})

	app.Command("install", "Install (or update) the project dependencies (docker containers, npm, composer...)", func(cmd *cli.Cmd) {
		cmd.Action = func() {

			action := config.Default.SrcPrepare
			fmt.Printf("\n ▶ ️ Prepare config files...\n")

			for _, commandDefinition := range action.Commands {
				cmd := NewCommand(commandDefinition)
				cmd.Execute()
			}

			action = config.Default.Install
			fmt.Printf("\n ▶ ️ Install the project...\n")

			for _, commandDefinition := range action.Commands {
				cmd := NewCommand(commandDefinition)
				cmd.Execute()
			}
		}
	})

	app.Command("custom", "Execute the custom actions", func(cmd *cli.Cmd) {
		cmd.Action = func() {

			for _, action := range config.Custom {

				fmt.Printf("\n ▶ ️ Executing [%s]\n", action.Name)

				for _, commandDefinition := range action.Commands {
					cmd := NewCommand(commandDefinition)
					cmd.Execute()
				}
			}

		}
	})

	app.Version("v version", "Pliz 0.1")

	app.Run(os.Args)
}
