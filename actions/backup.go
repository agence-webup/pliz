package actions

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"webup/pliz/config"
	"webup/pliz/domain"
	"webup/pliz/utils"

	"github.com/Songmu/prompter"
	"github.com/fatih/color"
	"github.com/jhoonb/archivex"
)

func BackupActionHandler(ctx domain.ExecutionContext, backupFilesOpt *bool, backupDBOpt *bool, outputOpt *string) {

	backupFiles := false
	if backupFilesOpt == nil && len(config.Get().BackupConfig.Files) > 0 {
		backupFiles = prompter.YN("Backup files?", true)
	} else if backupFilesOpt != nil {
		backupFiles = *backupFilesOpt
	}

	backupDB := false
	if backupDBOpt == nil && len(config.Get().BackupConfig.Databases) > 0 {
		backupDB = prompter.YN("Backup databases?", true)
	} else if backupDBOpt != nil {
		backupDB = *backupDBOpt
	}

	fmt.Println("")

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

	// save the archive with the right name
	archiveFilename := ""
	if outputOpt != nil && *outputOpt != "" {
		archiveFilename = *outputOpt
	} else {
		now := time.Now().UTC()
		year, month, day := now.Date()
		hour, minutes, seconds := now.Clock()
		archiveFilename = fmt.Sprintf("backup-%d%02d%02d_%02d%02d%02d.tar.gz", year, month, day, hour, minutes, seconds)
	}
	os.Rename(path.Join(backupDir, "backup_archive.tar.gz"), archiveFilename)

	// clean tmp
	os.RemoveAll(backupDir)

	fmt.Printf("\n %s Done\n", color.GreenString("✓"))
}

func makeDump(ctx domain.ExecutionContext, dbBackup domain.DatabaseBackupConfig, backupDir string) {

	// fetch the container id for db
	containerID, err := utils.GetContainerID(dbBackup.Container, ctx)
	if err != nil {
		return
	}

	// get the container config
	config, err := utils.GetContainerConfig(containerID, ctx)
	if err != nil {
		return
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
		err := mysqlDump(path.Join(backupDir, "dump.sql"), containerID, config.Env)
		if err != nil {
			fmt.Println(err)
		}
	} else if dbType == "mongo" {
		err := mongoDump(path.Join(backupDir, "mongodb.archive"), containerID, config.Env)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println("\nError: unsupported database (only MySQL or MongoDB)")
	}

}

func mysqlDump(destination string, containerId string, env domain.DockerContainerEnv) error {

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

func mongoDump(destination string, containerId string, env domain.DockerContainerEnv) error {

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
