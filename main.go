package main

import (
	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"

	"webup/pliz/actions"
	"webup/pliz/config"
	"webup/pliz/domain"
	"webup/pliz/utils"

	"github.com/Songmu/prompter"
	"github.com/fatih/color"
	cli "github.com/jawher/mow.cli"
)

//go:embed resources/*
var resources embed.FS

func main() {

	app := cli.App("pliz", "Manage projects building")

	app.Version("v version", "Pliz 11")

	// option to change the Pliz env
	plizEnv := app.String(cli.StringOpt{
		Name:  "env",
		Value: "",
		Desc:  "Change the environnment of Pliz (i.e. 'prod'). The environment var 'PLIZ_ENV' can be use too.",
	})
	prod := false
	var executionContext domain.ExecutionContext

	app.Before = func() {
		// Parse and check config
		parseAndCheckConfig()

		// check for env
		if *plizEnv == "prod" || os.Getenv("PLIZ_ENV") == "prod" {
			prod = true
		}

		executionContext = domain.ExecutionContext{Env: *plizEnv}
	}

	app.Command("start", "Start (or restart) the project", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			actions.StartActionHandler(prod, true)

			if !prod {
				// display access infos
				containerID, _ := utils.GetContainerID(config.Get().StartupContainer, executionContext)
				ports := utils.GetExposedPorts(containerID, executionContext)

				if len(ports) > 0 {
					// get ip from DOCKER_HOST env variable
					rgexp := regexp.MustCompile("(\\d{1,3}(?:\\.\\d{1,3}){3})")
					ip := rgexp.FindString(os.Getenv("DOCKER_HOST"))
					if ip == "" {
						ip = "localhost"
					}

					fmt.Printf("\nYour app is accessible using:\n")
					for _, port := range ports {
						color.Green("   http://%s:%s", ip, port)
					}
				} else {
					fmt.Printf("\n%s: The proxy doesn't seem to be exposed. Check your ports settings.\n", color.YellowString("Warning"))
				}
			}
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
		skipped := cmd.Strings(cli.StringsOpt{
			Name:   "s skip",
			Value:  []string{},
			Desc:   "Allows to skip some default tasks",
			EnvVar: "PLIZ_INSTALL_SKIP",
		})
		// cmd.StringsOpt("skip", []string{}, "")

		cmd.Action = func() {

			config := config.Get()

			if prod {
				backup := prompter.YN("You're in production. Do you want to make a backup?", true)
				if backup {
					err := actions.BackupActionHandler(executionContext, nil, nil, nil, nil)
					if err != nil {
						fmt.Printf("\n%s: %v\n", color.RedString("Error during backup"), err)
					}
					fmt.Println("")
				}

				ok := prompter.YN("The installation is going to start. Are you sure you want to continue?", false)
				if !ok {
					return
				}
			}

			/*
			 * 1. Duplicate and edit the config files (.env, docker_ports.yml...)
			 */

			fmt.Printf("\n %s ️ Prepare config files...\n\n", color.YellowString("▶"))

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

				fmt.Println(configFile.Target + color.GreenString(" OK."))
			}

			fmt.Println("")

			/*
			 * 2. Build the containers
			 */

			fmt.Printf("\n %s ️ Build the containers...\n", color.YellowString("▶"))

			cmd := domain.NewComposeCommand([]string{"build"}, prod)
			cmd.Execute()

			fmt.Println("")

			/*
			 * 3. Start the containers
			 */

			fmt.Printf("\n %s ️ Starting containers...\n", color.YellowString("▶"))

			// and start the containers
			actions.StartActionHandler(prod, false)

			fmt.Println("")

			/*
			 * 4. Run the enabled tasks
			 */

			fmt.Printf("\n %s ️ Run install tasks...\n", color.YellowString("▶"))

		TaskLoop:
			for _, id := range config.InstallTasks {

				task := config.Tasks[id]

				// check if the task is skipped
				for _, taskName := range *skipped {
					if taskName == string(task.Name) {
						fmt.Printf("\n%s %s %s\n", color.YellowString("-->"), task.Name, color.YellowString("skipped"))
						continue TaskLoop
					}
				}

				fmt.Printf("\n%s %s %s\n", color.CyanString("***"), task.Name, color.CyanString("***"))

				// disable the execution check if the installation is forced
				if *forced {
					task.ExecutionCheck = nil
				}

				if task.Execute(domain.TaskExecutionContext{Prod: prod}) {
					fmt.Printf("Task '%s' %s.\n", task.Name, color.GreenString("executed"))
				}
			}

			/*
			 * 5. Plug Git precommit Hook
			 */

			cmd = domain.NewCommand([]string{"pliz", "git:plug-hook", "-i"})
			cmd.Execute()
			/*
			 * 6. The end
			 */

			fmt.Printf("\n\n%s You may now run '%s' to launch your project\n", color.GreenString("✓"), color.MagentaString("pliz start"))

			if len(config.Checklist) > 0 {
				for _, item := range config.Checklist {
					fmt.Printf("  %s %s\n", color.RedString("→"), item)
				}
			}

			fmt.Println("")

		}
	})

	app.Command("bash", "Display a shell inside the builder service (or the specified service)", func(cmd *cli.Cmd) {

		// parse and check config
		parseAndCheckConfig()

		defaultContainer := config.Get().Containers.Builder

		cmd.Spec = "[-p...] [SERVICE]"

		container := cmd.StringArg("SERVICE", defaultContainer, "The Compose service that will be used to display the shell")
		ports := cmd.StringsOpt("p port", []string{}, "List of ports that will be published (e.g. 9000:3306)")

		cmd.Action = func() {

			options := []string{}
			if len(*ports) > 0 {
				for _, port := range *ports {
					options = append(options, "-p", port)
				}
			}

			cmd := domain.NewContainerCommand(*container, []string{"bash"}, options, prod)
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

	app.Command("run", "Execute a single task", func(cmd *cli.Cmd) {

		// parse and check config
		parseAndCheckConfig()

		taskIDs := []string{}
		for id := range config.Get().Tasks {
			taskIDs = append(taskIDs, string(id))
		}
		sort.Strings(taskIDs)

		for _, id := range taskIDs {
			task := config.Get().Tasks[domain.TaskID(id)]

			cmd.Command(id, task.Description, func(cmd *cli.Cmd) {
				cmd.Action = actions.RunTaskActionHandler(task, prod)
			})
		}
	})

	app.Command("backup", "Perform a backup of the project", func(cmd *cli.Cmd) {

		cmd.Spec = "[-q [--files] [--db]] [-o] [-k]"

		quiet := cmd.BoolOpt("q quiet", false, "Avoid prompt")
		backupFiles := cmd.BoolOpt("files", false, "Indicates if files will be backup")
		backupDB := cmd.BoolOpt("db", false, "Indicates if DB will be backup")

		outputFilename := cmd.StringOpt("o output", "", "Set the filename of the tar.gz")
		key := cmd.StringOpt("k", "", "the encryption password (32 characters for a key of 256 bits)")

		cmd.Action = func() {
			if *quiet == false {
				backupFiles = nil
				backupDB = nil
			}

			err := actions.BackupActionHandler(executionContext, backupFiles, backupDB, outputFilename, key)
			if err != nil {
				fmt.Printf("\n%s: %v\n", color.RedString("Error during backup"), err)
				cli.Exit(1)
			}
		}
	})

	app.Command("restore", "Restore a backup (Warning: files will be overrided)", func(cmd *cli.Cmd) {

		cmd.Spec = "[-q [--config-files] [--files] [--db]] [-k] FILE"

		quiet := cmd.BoolOpt("q quiet", false, "Avoid prompt")
		restoreConfigFiles := cmd.BoolOpt("config-files", false, "Indicates if config files will be restored")
		restoreFiles := cmd.BoolOpt("files", false, "Indicates if files will be restored")
		restoreDB := cmd.BoolOpt("db", false, "Indicates if DB will be restored")
		key := cmd.StringOpt("k", "", "the encryption password")

		file := cmd.StringArg("FILE", "", "A pliz backup file (tar.gz)")

		cmd.Action = func() {
			if *quiet == false {
				restoreConfigFiles = nil
				restoreFiles = nil
				restoreDB = nil
			}

			actions.RestoreActionHandler(executionContext, *file, restoreConfigFiles, restoreFiles, restoreDB, key)
		}
	})

	app.Command("git:plug-hook", "Register hooks in the Git repository", func(cmd *cli.Cmd) {
		include := cmd.BoolOpt("i include-phpcs-config", false, "Includes the default phpcs config, if not already setup in the repo")
		force := cmd.BoolOpt("f force", false, "Forces the include of the default phpcs config, even if already setup in the repo")

		cmd.Action = func() {
			fmt.Printf("%s git:plug-hook\n\n", color.MagentaString("Executing:"))

			plugPhpcsConfig(*include, *force)
			plugHook()
		}
	})

	app.Run(os.Args)
}

func plugPhpcsConfig(include bool, force bool) {
	if !include {
		return
	}

	_, err := ioutil.ReadFile(".phpcs.xml.dist")

	// if file already exists and there is no force flag: return
	if err == nil && !force {
		fmt.Printf("%s .phpcs.xml.dist already exists.\n", color.YellowString("Skipping include:"))
		return
	}

	config, err := resources.ReadFile("resources/.phpcs.xml.dist")
	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Unable to read default config:"), err)
		return
	}

	// write +wr for everybody
	if err := ioutil.WriteFile(".phpcs.xml.dist", config, 0555); err != nil {
		fmt.Printf("%s %v\n", color.RedString("Unable to write file:"), err)
		return
	}

	fmt.Printf("%s ./.phpcs.xml.dist\n", color.GreenString("Successfully plugged:"))
}

func plugHook() {
	// open the hook code from the project config
	script, err := ioutil.ReadFile("ops/git/hooks/pre-commit")

	// if it cannot be read, plug the default hook instead
	if err != nil {
		plugDefaultHook()
		return
	}
	plugProjectHook(script)
}

func plugProjectHook(script []byte) {
	// write +wr for everybody, +x for user
	err := ioutil.WriteFile(".git/hooks/pre-commit", script, 0755)

	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Unable to write file:"), err)
		return
	}
	fmt.Printf("%s .git/hooks/pre-commit\n", color.GreenString("Successfully plugged:"))
}

func plugDefaultHook() {
	fmt.Printf("./ops/git/hooks/pre-commit not found, fallback to default hook\n")

	script, err := resources.ReadFile("resources/git-hooks/pre-commit")

	if err != nil {
		fmt.Printf("%s %v\n", color.RedString("Unable to read default hook:"), err)
		return
	}

	if err := ioutil.WriteFile(".git/hooks/pre-commit", script, 0755); err != nil {
		fmt.Printf("%s %v\n", color.RedString("Unable to write file:"), err)
		return
	}

	fmt.Printf("%s .git/hooks/pre-commit\n", color.GreenString("Successfully plugged:"))
}

func parseAndCheckConfig() {
	err := config.Check()
	if err != nil {
		os.Exit(1)
		return
	}
}
