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

func BackupActionHandler(ctx domain.ExecutionContext, backupFilesOpt *bool, backupDBOpt *bool, outputOpt *string, key *string, verbose bool) error {

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
	backupDir := ".pliz_backup"
	err := os.Mkdir(backupDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Unable to create a backup directory: %s\n", err)
	}
	defer os.RemoveAll(backupDir)

	// config files backup
	if len(config.Get().ConfigFiles) > 0 {
		dir := path.Join(backupDir, "backup", "config")
		os.MkdirAll(dir, 0755)
		for _, configFile := range config.Get().ConfigFiles {
			if _, err = os.Stat(configFile.Target); err == nil {
				target := path.Join(dir, configFile.Target)
				os.MkdirAll(filepath.Dir(target), os.ModePerm)
				os.Link(configFile.Target, target)
			} else {
				return fmt.Errorf("Unable to backup a config file: %s\n", err)
			}
		}
	}

	if backupDB {
		// databases dump
		for _, dbBackup := range config.Get().BackupConfig.Databases {
			dir := path.Join(backupDir, "backup", "databases", dbBackup.Container)
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return fmt.Errorf("Unable to create the db backup directory: %s\n", err)
			}
			err = makeDump(ctx, dbBackup, dir, verbose)
			if err != nil {
				return fmt.Errorf("Unable to backup databases: %s\n", err)
			}
		}
	}

	if backupFiles {
		// files
		filesDir := path.Join(backupDir, "backup", "files")
		os.MkdirAll(filesDir, os.ModePerm)

		// prepare a walk function to handle the whole hierarchy
		walkFunc := func(filepath string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("%s file or directory not found \n%s\n", filepath, err)
			}
			target := path.Join(filesDir, filepath)

			// just create the directory
			if info.IsDir() {
				err := os.MkdirAll(target, os.ModePerm)
				return err
			}

			// perform a hard link
			linkErr := os.Link(filepath, target)
			return linkErr
		}

		for _, file := range config.Get().BackupConfig.Files {
			err = filepath.Walk(file, walkFunc)
			if err != nil {
				return fmt.Errorf("Unable to walk into %s\n%s\n", file, err)
			}
		}
	}

	tmpArchiveFilename := path.Join(backupDir, "backup_archive.tar.gz")

	tar := new(archivex.TarFile)
	tar.Create(tmpArchiveFilename)
	tar.AddAll(path.Join(backupDir, "backup"), false)
	tar.Close()

	if key != nil && *key != "" {
		tmpEncryptedFilename := path.Join(backupDir, "backup_archive.tar.gz.enc")
		infile, err := os.Open(tmpArchiveFilename)
		if err != nil {
			return fmt.Errorf("Unable to open the archive: %s\n", err)
		}

		outfile, err := os.OpenFile(tmpEncryptedFilename, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("Unable to create the encrypted file: %s\n", err)
		}

		err = utils.Encrypt(infile, outfile, []byte(*key))
		if err != nil {
			return fmt.Errorf("Unable to encrypt file: %s\n", err)
		}
		infile.Close()
		outfile.Close()

		tmpArchiveFilename = tmpEncryptedFilename
	}

	// save the archive with the right name
	archiveFilename := ""
	if outputOpt != nil && *outputOpt != "" {
		archiveFilename = *outputOpt
	} else {
		now := time.Now().UTC()
		year, month, day := now.Date()
		hour, minutes, seconds := now.Clock()
		encryptedExtension := ""
		if key != nil && *key != "" {
			encryptedExtension = ".enc"
		}
		archiveFilename = fmt.Sprintf("backup-%d%02d%02d_%02d%02d%02d.tar.gz%s", year, month, day, hour, minutes, seconds, encryptedExtension)
	}

	err = os.Rename(tmpArchiveFilename, archiveFilename)
	if err != nil {
		return fmt.Errorf("Unable to create the backup file: %s\n", err)
	}

	// clean tmp
	// err = os.RemoveAll(backupDir)
	// if err != nil {
	// 	return fmt.Errorf("Unable to remove temp folder: %s\n", err)
	// }

	fmt.Printf("\n %s Done\n", color.GreenString("✓"))
	return nil
}

func makeDump(ctx domain.ExecutionContext, dbBackup domain.DatabaseBackupConfig, backupDir string, verbose bool) error {

	// fetch the container id for db
	containerID, err := utils.GetContainerID(dbBackup.Container, ctx)
	if err != nil {
		return err
	}

	// get the container config
	config, err := utils.GetContainerConfig(containerID, ctx)
	if err != nil {
		return err
	}

	// get the type of DB to backup
	// check from config or try to find it with the image name
	dbType := dbBackup.Type
	if dbType == "" {
		if strings.Contains(config.Image, "mysql") {
			dbType = "mysql"
		} else if strings.Contains(config.Image, "mariadb") {
			dbType = "mariadb"
		} else if strings.Contains(config.Image, "postgres") {
			dbType = "postgres"
		} else if strings.Contains(config.Image, "mongo") {
			dbType = "mongo"
		}
	}

	if dbType == "mysql" || dbType == "mariadb" {
		return mysqlDump(containerID, config.Env, backupDir, dbBackup.Databases, dbBackup.NoLock, verbose)
	} else if dbType == "postgres" {
		return postgresDump(containerID, config.Env, backupDir, dbBackup.Databases, verbose)
	} else if dbType == "mongo" {
		return mongoDump(path.Join(backupDir, "mongodb.archive"), containerID, config.Env, backupDir)
	} else {
		return fmt.Errorf("\nError: unsupported database (only MySQL, MariaDB, PostgreSQL or MongoDB)")
	}
}

func mysqlDump(containerId string, env domain.DockerContainerEnv, backupDir string, databases []string, noLock bool, verbose bool) error {

	password := ""
	if value, ok := env["MYSQL_ROOT_PASSWORD"]; ok {
		password = value
	}

	mysqlDatabases := []string{"db"}
	if len(databases) > 0 {
		mysqlDatabases = databases
	} else {
		if value, ok := env["MYSQL_DATABASE"]; ok {
			mysqlDatabases = []string{value}
		}
	}

	for _, database := range mysqlDatabases {
		cmdArgs := []string{"docker", "exec", "-i", containerId, "mysqldump", fmt.Sprintf("--password=%s", password), database}
		if noLock {
			cmdArgs = []string{"docker", "exec", "-i", containerId, "mysqldump", "--single-transaction", "--skip-lock-tables", fmt.Sprintf("--password=%s", password), database}
		}

		cmd := domain.NewCommand(cmdArgs, verbose)

		file, err := ioutil.TempFile(backupDir, "plizdump")
		if err != nil {
			fmt.Println("Unable to create a tmp file:")
			return err
		}
		defer file.Close()

		if err := cmd.WriteResultToFile(file); err != nil {
			os.Remove(file.Name())
			return err
		}

		os.Rename(file.Name(), path.Join(backupDir, database+".sql"))
	}

	return nil
}

func postgresDump(containerId string, env domain.DockerContainerEnv, backupDir string, databases []string, verbose bool) error {

	password := ""
	if value, ok := env["POSTGRES_PASSWORD"]; ok {
		password = value
	}

	user := "postgres"
	if value, ok := env["POSTGRES_USER"]; ok {
		user = value
	}

	postgresDatabases := []string{"db"}
	if len(databases) > 0 {
		postgresDatabases = databases
	} else {
		if value, ok := env["POSTGRES_DB"]; ok {
			postgresDatabases = []string{value}
		}
	}

	for _, database := range postgresDatabases {
		cmd := domain.NewCommand([]string{"docker", "exec", "-i", "-e", fmt.Sprintf("PGPASSWORD=\"%s\"", password), containerId, "pg_dump", "-Fc", fmt.Sprintf("--username=%s", user), database}, verbose)

		file, err := ioutil.TempFile(backupDir, "plizdump")
		if err != nil {
			fmt.Println("Unable to create a tmp file:")
			return err
		}
		defer file.Close()

		if err := cmd.WriteResultToFile(file); err != nil {
			os.Remove(file.Name())
			return err
		}

		os.Rename(file.Name(), path.Join(backupDir, database+".dump"))
	}

	return nil
}

func mongoDump(destination string, containerId string, env domain.DockerContainerEnv, backupDir string) error {

	cmd := domain.NewCommand([]string{"docker", "exec", "-i", containerId, "mongodump", "--archive", "--gzip"}, true)

	file, err := ioutil.TempFile(backupDir, "plizdump")
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
