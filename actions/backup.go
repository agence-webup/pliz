package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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

		backupFiles := false
		if len(config.Get().BackupConfig.Files) > 0 {
			backupFiles = prompter.YN("Backup files?", true)
		}

		backupDB := false
		if len(config.Get().BackupConfig.Databases) > 0 {
			backupDB = prompter.YN("Backup databases?", true)
		}

		// prepare the directory to store the backup
		backupDir, err := ioutil.TempDir("", "plizbackup")
		if err != nil {
			fmt.Printf("Unable to create a backup directory: %s\n", err)
			return
		}

		// config files backup
		if len(config.Get().ConfigFiles) > 0 {
			dir := path.Join(backupDir, "backup", "config")
			os.MkdirAll(dir, 0755)
			for _, configFile := range config.Get().ConfigFiles {
				if _, err := os.Stat(configFile.Target); err == nil {
					target := path.Join(dir, configFile.Target)
					os.MkdirAll(filepath.Dir(target), 0755)
					os.Link(configFile.Target, target)
				} else {
					fmt.Println(err)
				}
			}
		}

		if backupDB {
			// databases dump
			for _, dbBackup := range config.Get().BackupConfig.Databases {
				dir := path.Join(backupDir, "backup", "databases", dbBackup.Container)
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					fmt.Println(err)
					return
				}
				makeDump(ctx, dbBackup, dir)
			}
		}

		if backupFiles {
			// files
			filesDir := path.Join(backupDir, "backup", "files")
			os.MkdirAll(filesDir, 0755)
			for _, file := range config.Get().BackupConfig.Files {
				target := path.Join(filesDir, file)
				os.MkdirAll(filepath.Dir(target), 0755)
				os.Link(file, target)
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

func makeDump(ctx domain.ExecutionContext, dbBackup domain.DatabaseBackupConfig, backupDir string) {

	// fetch the container id for db
	cmd := domain.NewComposeCommand([]string{"ps", "-q", dbBackup.Container}, ctx.IsProd())
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

	// get the type of DB to backup
	// check from config or try to find it with the image name
	dbType := dbBackup.Type
	if dbType == "" {
		if strings.Contains(config.Image, "mysql") {
			dbType = "mysql"
		} else if strings.Contains(config.Image, "mongo") {
			dbType = "mongo"
		}
	}

	if dbType == "mysql" {
		fmt.Println(path.Join(backupDir, "dump.sql"))
		err := mysqlDump(path.Join(backupDir, "dump.sql"), containerId, env)
		if err != nil {
			fmt.Println(err)
		}
	} else if dbType == "mongo" {
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

	cmd := domain.NewCommand([]string{"docker", "exec", "-i", containerId, "mongodump", "--archive", "--gzip"})

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
