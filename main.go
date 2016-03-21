package main

import (
	"fmt"
	"os"
	"webup/pliz/config"
	"webup/pliz/domain"
	"webup/pliz/tasks"
	"webup/pliz/utils"

	"github.com/Songmu/prompter"
	"github.com/jawher/mow.cli"
)

func main() {

	app := cli.App("pliz", "Manage projects building")

	app.Version("v version", "Pliz 2-dev (build 5)")

	// option to change the Pliz env
	plizEnv := app.String(cli.StringOpt{
		Name:  "env",
		Value: "",
		Desc:  "Change the environnment of Pliz (i.e. 'prod'). The environment var 'PLIZ_ENV' can be use too.",
	})
	prod := false

	app.Before = func() {
		// Parse and check config
		parseAndCheckConfig()

		// check for env
		if *plizEnv == "prod" || os.Getenv("PLIZ_ENV") == "prod" {
			prod = true
		}
	}

	app.Command("start", "Start (or restart) the project", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			cmd := domain.NewComposeCommand([]string{"up", "-d", config.Get().Containers.Proxy}, prod)
			cmd.Execute()
		}
	})

	app.Command("stop", "Stop the project", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			cmd := domain.NewComposeCommand([]string{"stop"}, prod)
			cmd.Execute()
		}
	})

	app.Command("install", "Install (or update) the project dependencies (docker containers, npm, composer...)", func(cmd *cli.Cmd) {

		forced := cmd.BoolOpt("f force", false, "Force the installation process")

		cmd.Action = func() {

			config := config.Get()

			if prod {
				backup := prompter.YN("You're in production. Do you want to make a backup?", true)
				fmt.Println("Backup:", backup)

				ok := prompter.YN("The installation is going to start. Are you sure you want to continue?", false)
				if !ok {
					return
				}
			}

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
					utils.CopyFileContents(configFile.Sample, configFile.Target)
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

			cmd := domain.NewComposeCommand([]string{"build"}, prod)
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

				if task.Execute(domain.TaskExecutionContext{Prod: prod}) {
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

	app.Command("bash", "Display a shell inside the builder service (or the specified service)", func(cmd *cli.Cmd) {

		// parse and check config
		parseAndCheckConfig()

		defaultContainer := config.Get().Containers.Builder

		cmd.Spec = "[SERVICE]"
		container := cmd.StringArg("SERVICE", defaultContainer, "The Compose service that will be used to display the shell")

		cmd.Action = func() {
			cmd := domain.NewContainerCommand(*container, []string{"bash"}, prod)
			cmd.Execute()
		}
	})

	app.Command("logs", "Display logs of all services (or the specified service)", func(cmd *cli.Cmd) {

		cmd.Spec = "[SERVICE]"
		container := cmd.StringArg("SERVICE", "", "The Compose service to log")

		cmd.Action = func() {

			args := []string{"logs"}

			if *container != "" {
				args = append(args, *container)
			}

			cmd := domain.NewComposeCommand(args, prod)
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

			if task.Execute(domain.TaskExecutionContext{Prod: prod}) {
				fmt.Printf("Task '%s' executed.\n", task.Name)
			}
		}
	})

	app.Run(os.Args)
}

func parseAndCheckConfig() {
	err := config.Check()
	if err != nil {
		os.Exit(1)
		return
	}
}
