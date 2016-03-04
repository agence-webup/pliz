package main

import (
	"fmt"
	"os"
	"webup/pliz/config"
	"webup/pliz/domain"
	"webup/pliz/tasks"

	"github.com/jawher/mow.cli"
)

func main() {

	// Parse and check config
	err := config.Check()
	if err != nil {
		os.Exit(1)
		return
	}

	app := cli.App("pliz", "Manage projects building")

	app.Command("ps", "List running containers", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			command := domain.Command{Name: "docker", Args: []string{"ps"}}
			command.Execute()
		}
	})

	app.Command("config", "...", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			fmt.Println(config.Get())
		}
	})

	app.Command("install", "Install (or update) the project dependencies (docker containers, npm, composer...)", func(cmd *cli.Cmd) {

		forced := cmd.BoolOpt("f force", false, "Force the installation process")

		cmd.Action = func() {

			config := config.Get()

			/*
			 * 1. Duplicate and edit the config files (.env, docker_ports.yml...)
			 */

			fmt.Printf("\n ▶ ️ Prepare config files...\n\n")

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
					cmd := domain.NewCommand([]string{"vim", configFile.Target})
					cmd.Execute()
				}

				fmt.Println(configFile.Target + " OK.")
			}

			fmt.Println("")

			/*
			 * 2. Build the containers
			 */

			fmt.Printf("\n ▶ ️ Build the containers...\n")

			cmd := domain.NewCommand([]string{"docker-compose", "build"})
			cmd.Execute()

			fmt.Println("")

			/*
			 * 3. Run the enabled tasks
			 */

			fmt.Printf("\n ▶ ️ Run enabled tasks...\n")

			for _, task := range config.EnabledTasks {
				fmt.Println("\n*** " + task.Name + " ***")
				// task, err := tasks.CreateTaskWithName(taskName)

				// if err != nil {
				// 	fmt.Println(err)
				// 	continue
				// }

				if task.Execute() {
					fmt.Printf("Task '%s' executed.\n", task.Name)
				} else {
					// fmt.Printf("Task '%s' skipped.\n", task.Name)
					// fmt.Println(err)
				}
			}

			action := config.Default.SrcPrepare
			for _, commandDefinition := range action.Commands {
				cmd := domain.NewCommand(commandDefinition)
				cmd.Execute()
			}

			action = config.Default.Install
			fmt.Printf("\n ▶ ️ Install the project...\n")

			for _, commandDefinition := range action.Commands {
				cmd := domain.NewCommand(commandDefinition)
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

	app.Command("bash", "Display a shell inside the build container (or the specified container)", func(cmd *cli.Cmd) {

		defaultContainer := config.Get().Containers.Builder

		cmd.Spec = "[CONTAINER]"
		container := cmd.StringArg("CONTAINER", defaultContainer, "The container that will be used to display the shell")

		cmd.Action = func() {

			// NOTE: this code allows to run a shell inside a running container. Not used currently.

			// cmd1 := exec.Command("docker-compose", "ps", "-q", *container)
			// output, err := cmd1.Output()
			// if err != nil {
			// 	fmt.Println(err)
			// 	cli.Exit(1)
			// 	return
			// }
			// containerId := strings.TrimSpace(string(output))
			// if containerId == "" {
			// 	fmt.Printf("The container '%s' is not running.", *container)
			// 	cli.Exit(1)
			// 	return
			// }
			//
			// cmd := domain.NewCommand([]string{
			// 	"docker",
			// 	"exec",
			// 	"-it",
			// 	containerId,
			// 	"bash",
			// })

			cmd := domain.NewContainerCommand(*container, []string{"bash"})
			cmd.Execute()
		}
	})

	app.Command("run", "Execute a task", func(cmd *cli.Cmd) {

		cmd.Spec = "TASK"
		taskName := cmd.StringArg("TASK", "", "Execute the specified task. Run 'pliz tasks' to get the list of the tasks")

		cmd.Action = func() {

			// get the task
			task, err := tasks.CreateTaskWithName(*taskName, config.Get())
			if err != nil {
				fmt.Println(err)
				cli.Exit(1)
				return
			}

			// disable the execution check for standalone execution
			task.ExecutionCheck = nil

			if task.Execute() {
				fmt.Printf("Task '%s' executed.\n", task.Name)
			}
		}
	})

	app.Command("custom", "Execute the custom actions", func(cmd *cli.Cmd) {
		cmd.Action = func() {

			config := config.Get()

			for _, action := range config.Custom {

				fmt.Printf("\n ▶ ️ Executing [%s]\n", action.Name)

				for _, commandDefinition := range action.Commands {
					cmd := domain.NewCommand(commandDefinition)
					cmd.Execute()
				}
			}

		}
	})

	app.Version("v version", "Pliz 0.1")

	app.Run(os.Args)
}
