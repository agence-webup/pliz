package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"webup/pliz/config"
	"webup/pliz/domain"
)

type containerConfig struct {
	Env   []string
	Image string
}

type containerEnv map[string]string

func BackupActionHandler(ctx domain.ExecutionContext) func() {
	return func() {

		if len(config.Get().BackupConfig.Databases) == 0 && len(config.Get().BackupConfig.Files) == 0 {
			fmt.Println("'backup' section is empty or not defined in the config file")
			return
		}

		// prepare the directory to store the backup
		backupDir, err := ioutil.TempDir("", "plizbackup")
		if err != nil {
			fmt.Printf("Unable to create a backup directory: %s\n", err)
			return
		}

		// dump
		for _, dbService := range config.Get().BackupConfig.Databases {
			makeDump(ctx, dbService, backupDir)
		}

		// if err := cmd.WriteResultToFile(file); err != nil {
		// 	os.Remove(file.Name())
		// 	return err
		// }

		os.Rename(backupDir, "backup-test")

	}
}

func makeDump(ctx domain.ExecutionContext, dbContainer string, backupDir string) {

	// fetch the container id for db
	cmd := domain.NewComposeCommand([]string{"ps", "-q", dbContainer}, ctx.IsProd())
	containerId, err := cmd.GetResult()
	if err != nil {
		fmt.Println("Unable to get the 'db' container id")
	}

	fmt.Println(containerId)

	// get the container config
	cmd = domain.NewCommand([]string{"docker", "inspect", "--format", "{{json .Config}}", containerId})
	configJson, err := cmd.GetResult()
	if err != nil {
		fmt.Println("Unable to get the config of the 'db' container")
	}

	// parse the json
	var config containerConfig
	json.NewDecoder(strings.NewReader(configJson)).Decode(&config)

	// parse env variables of the container
	env := containerEnv{}
	for _, data := range config.Env {
		items := strings.SplitN(data, "=", 2)
		env[items[0]] = items[1]
	}

	if strings.Contains(config.Image, "mysql") {
		err := mysqlDump(path.Join(backupDir, "dump.sql"), containerId, env)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func mysqlDump(destination string, containerId string, env containerEnv) error {

	password := ""
	if value, ok := env["MYSQL_ROOT_PASSWORD"]; ok {
		password = value
	}

	database := ""
	if value, ok := env["MYSQL_DATABASE"]; ok {
		database = value
	}

	cmd := domain.NewCommand([]string{"docker", "exec", "-i", containerId, "mysqldump", fmt.Sprintf("--password=%s", password), database})

	file, err := ioutil.TempFile("", "plizdump")
	if err != nil {
		fmt.Println("Unable to create a tmp file:")
		return err
	}
	defer file.Close()

	if err := cmd.WriteResultToFile(file); err != nil {
		os.Remove(file.Name())
		return err
	}

	os.Rename(file.Name(), destination)

	return nil
}
