package main

import (
	"fmt"
	"os"
	"webup/pliz/helpers"
	"webup/pliz/tasks"

	"github.com/jawher/mow.cli"
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

func main() {
	app := cli.App("pliz", "Manage projects building")

	var config Config

	app.Before = func() {
		err := GetConfig(&config)
		if err != nil {
			cli.Exit(1)
		}
	}

	app.Command("ps", "List running containers", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			command := helpers.Command{Name: "docker", Args: []string{"ps"}}
			command.Execute()
		}
	})

	app.Command("config", "...", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			fmt.Println(config)
		}
	})

	app.Command("install", "Install (or update) the project dependencies (docker containers, npm, composer...)", func(cmd *cli.Cmd) {

		forced := cmd.BoolOpt("f force", false, "Force the installation process")

		cmd.Action = func() {

			/*
			 * 1. Duplicate and edit the config files (.env, docker_ports.yml...)
			 */

			fmt.Printf("\n ▶ ️ Prepare config files...\n")

			for _, configFile := range config.ConfigFiles {
				created := false

				// if the 'force' option is set, then we consider that the file has been created
				if *forced {
					created = true
				}

				// check if the file exists. If not, duplicate the sample
				if _, err := os.Stat(configFile.Target); os.IsNotExist(err) {
					os.Link(configFile.Sample, configFile.Target)
					created = true
				}

				// edit the file
				if created {
					cmd := helpers.NewCommand([]string{"vim", configFile.Target})
					cmd.Execute()
				}
			}

			/*
			 * 2. Run the enabled tasks
			 */

			fmt.Printf("\n ▶ ️ Run enabled tasks...\n\n")

			for _, taskName := range config.EnabledTasks {
				if taskName == "npm" {
					task := tasks.NpmTask()

					if task.Execute() {
						fmt.Printf("Task '%s' executed.\n", task.Name)
					} else {
						// fmt.Printf("Task '%s' skipped.\n", task.Name)
						// fmt.Println(err)
					}
				}
			}

			action := config.Default.SrcPrepare
			for _, commandDefinition := range action.Commands {
				cmd := helpers.NewCommand(commandDefinition)
				cmd.Execute()
			}

			action = config.Default.Install
			fmt.Printf("\n ▶ ️ Install the project...\n")

			for _, commandDefinition := range action.Commands {
				cmd := helpers.NewCommand(commandDefinition)
				cmd.Execute()
			}
		}
	})

	app.Command("tasks", "Describe the available tasks", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			tasks := []string{
				"npm",
				"bower",
				"composer",
				"gulp",
				"db-update",
			}

			fmt.Println("Available tasks:")
			for _, task := range tasks {
				fmt.Printf("   %s\n", task)
			}
		}
	})

	app.Command("custom", "Execute the custom actions", func(cmd *cli.Cmd) {
		cmd.Action = func() {

			for _, action := range config.Custom {

				fmt.Printf("\n ▶ ️ Executing [%s]\n", action.Name)

				for _, commandDefinition := range action.Commands {
					cmd := helpers.NewCommand(commandDefinition)
					cmd.Execute()
				}
			}

		}
	})

	app.Version("v version", "Pliz 0.1")

	app.Run(os.Args)
}
