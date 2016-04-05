package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
	"webup/pliz/config"
	"webup/pliz/domain"

	"github.com/Songmu/prompter"
	"github.com/jhoonb/archivex"
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

		fmt.Println("What do you want to backup ?")
		backupFiles := prompter.YN("Files", true)
		backupDB := prompter.YN("DB", true)

		if !backupFiles && !backupDB {
			fmt.Println("Are you serious? Tsss...")
			return
		}

		// prepare the directory to store the backup
		backupDir, err := ioutil.TempDir("", "plizbackup")
		if err != nil {
			fmt.Printf("Unable to create a backup directory: %s\n", err)
			return
		}

		if backupDB {
			// databases dump
			for _, dbService := range config.Get().BackupConfig.Databases {
				dir := path.Join(backupDir, "backup", "databases", dbService)
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					fmt.Println(err)
					return
				}
				makeDump(ctx, dbService, dir)
			}
		}

		if backupFiles {
			// files
			filesDir := path.Join(backupDir, "backup", "files")
			os.MkdirAll(filesDir, 0755)
			for _, file := range config.Get().BackupConfig.Files {
				os.Link(file, path.Join(filesDir, file))
			}
		}

		// if err := cmd.WriteResultToFile(file); err != nil {
		// 	os.Remove(file.Name())
		// 	return err
		// }

		tar := new(archivex.TarFile)
		tar.Create(path.Join(backupDir, "backup_archive.tar.gz"))
		tar.AddAll(path.Join(backupDir, "backup"), false)
		tar.Close()

		now := time.Now().UTC()
		year, month, day := now.Date()
		hour, minutes, seconds := now.Clock()
		os.Rename(path.Join(backupDir, "backup_archive.tar.gz"), fmt.Sprintf("backup-%d%02d%02d_%02d%02d%02d.tar.gz", year, month, day, hour, minutes, seconds))
		os.RemoveAll(backupDir)

		fmt.Println("\n ✓ Done")
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
		fmt.Println(path.Join(backupDir, "dump.sql"))
		err := mysqlDump(path.Join(backupDir, "dump.sql"), containerId, env)
		if err != nil {
			fmt.Println(err)
		}
	} else if strings.Contains(config.Image, "mongo") {
		fmt.Println(path.Join(backupDir, "mongodb.archive"))
		err := mongoDump(path.Join(backupDir, "mongodb.archive"), containerId, env)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println("\nError: unsupported database (only MySQL or MongoDB)")
	}

}

func mysqlDump(destination string, containerId string, env containerEnv) error {

	password := ""
	if value, ok := env["MYSQL_ROOT_PASSWORD"]; ok {
		password = value
	}

	database := "db"
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

func mongoDump(destination string, containerId string, env containerEnv) error {

	cmd := domain.NewCommand([]string{"docker", "exec", "-i", containerId, "mongodump", "--archive"})

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
