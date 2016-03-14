package main

import (
	"fmt"
	"os"
	"webup/pliz/config"
	"webup/pliz/domain"
	"webup/pliz/tasks"

	"github.com/Songmu/prompter"
	"github.com/jawher/mow.cli"
)

func main() {

	app := cli.App("pliz", "Manage projects building")

	app.Version("v version", "Pliz 1.1-dev")

	app.Before = func() {
		// Parse and check config
		ParseAndCheckConfig()
	}

	app.Command("start", "Start the project", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			cmd := domain.NewCommand([]string{"docker-compose", "up", "-d", config.Get().Containers.Proxy})
			cmd.Execute()
		}
	})

	app.Command("stop", "Stop the project", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			cmd := domain.NewCommand([]string{"docker-compose", "stop"})
			cmd.Execute()
		}
	})

	app.Command("restart", "Restart the project", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			cmd := domain.NewCommand([]string{"docker-compose", "restart"})
			cmd.Execute()
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

				// disable the execution check if the installation is forced
				if *forced {
					task.ExecutionCheck = nil
				}

				if task.Execute() {
					fmt.Printf("Task '%s' executed.\n", task.Name)
				}
			}

			/*
			 * 4. The end
			 */

			fmt.Println("\n\n ✓ You may now run 'pliz start' to launch your project")

			if len(config.Checklist) > 0 {
				for _, item := range config.Checklist {
					fmt.Printf("  → %s\n", item)
				}
			}

			fmt.Println("")

		}
	})

	app.Command("bash", "Display a shell inside the build container (or the specified container)", func(cmd *cli.Cmd) {

		// parse and check config
		ParseAndCheckConfig()

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

	app.Command("logs", "Display logs of all containers (or the specified container)", func(cmd *cli.Cmd) {

		cmd.Spec = "[CONTAINER]"
		container := cmd.StringArg("CONTAINER", "", "The container to log")

		cmd.Action = func() {

			args := []string{"docker-compose", "logs"}

			if *container != "" {
				args = append(args, *container)
			}

			cmd := domain.NewCommand(args)
			cmd.Execute()
		}
	})

	app.Command("tasks", "List the available tasks", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			tasks := tasks.AllTaskNames()

			fmt.Println("Available tasks:")
			for _, task := range tasks {
				fmt.Printf("   %s\n", task)
			}
		}
	})

	app.Command("run", "Execute a task", func(cmd *cli.Cmd) {

		cmd.Spec = "TASK"
		taskName := cmd.StringArg("TASK", "", "Execute the specified task. Run 'pliz tasks' to get the list of the tasks")

		cmd.Action = func() {

			var task domain.Task

			// first, search for the enabled task (which could be overrided)
			taskFound := false
			for _, t := range config.Get().EnabledTasks {
				if *taskName == t.Name {
					task = t
					taskFound = true
					break
				}
			}

			// if no task is found, try to create a default task
			if !taskFound {
				// get the task
				defaultTask, err := tasks.CreateTaskWithName(*taskName, config.Get())
				if err != nil {
					fmt.Println(err)
					cli.Exit(1)
					return
				} else {
					task = defaultTask
				}
			}

			// disable the execution check for standalone execution
			task.ExecutionCheck = nil

			if task.Execute() {
				fmt.Printf("Task '%s' executed.\n", task.Name)
			}
		}
	})

	app.Command("deploy", "Execute deployment tasks", func(cmd *cli.Cmd) {

		cmd.Command("run", "Run a deployment", func(cmd *cli.Cmd) {

			cmd.Action = func() {
				backup := prompter.YN("Do you want to make a backup?", true)
				fmt.Println("Backup:", backup)

				ok := prompter.YN("Are you ready to deploy?", false)
				if !ok {
					return
				}

				cmd := domain.NewCommand([]string{"docker-compose", "-f", "docker-compose.yml", "-f", "docker-compose.prod.yml", "up", "-d", "proxy"})
				cmd.Execute()
			}

		})

	})

	app.Run(os.Args)
}

func ParseAndCheckConfig() {
	err := config.Check()
	if err != nil {
		os.Exit(1)
		return
	}
}
