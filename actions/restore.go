package actions

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/fatih/color"

	"webup/pliz/config"
	"webup/pliz/domain"
	"webup/pliz/utils"
)

// RestoreActionHandler handle the action for 'pliz restore'
func RestoreActionHandler(ctx domain.ExecutionContext, file string, restoreConfigFilesOpt *bool, restoreFilesOpt *bool, restoreDBOpt *bool, key *string) {

	isQuiet := !(restoreConfigFilesOpt == nil && restoreFilesOpt == nil && restoreDBOpt == nil)

	if ctx.IsProd() && !isQuiet {
		ok := prompter.YN("You're in production. Are you sure you want to continue?", false)
		if !ok {
			return
		}
	}

	if !isQuiet {
		fmt.Printf(" %s Choose what you want to restore:\n", color.YellowString("▶"))
	}

	configFilesRestoration := false
	if restoreConfigFilesOpt == nil && len(config.Get().ConfigFiles) > 0 {
		configFilesRestoration = prompter.YN("     - configuration files", false)
	} else if restoreConfigFilesOpt != nil {
		configFilesRestoration = *restoreConfigFilesOpt
	}

	filesRestoration := false
	if restoreFilesOpt == nil && len(config.Get().BackupConfig.Files) > 0 {
		filesRestoration = prompter.YN("     - others files", false)
	} else if restoreFilesOpt != nil {
		filesRestoration = *restoreFilesOpt
	}

	dbRestoration := false
	if restoreDBOpt == nil && len(config.Get().BackupConfig.Databases) > 0 {
		dbRestoration = prompter.YN("     - database dumps", false)
	} else if restoreDBOpt != nil {
		dbRestoration = *restoreDBOpt
	}

	fmt.Printf("\n\n")

	encryptedExtension := file[len(file)-4:]
	isEncrypted := encryptedExtension == ".enc" && key != nil || *key != ""

	if isEncrypted {
		var err error
		file, err = decrypt(file, key)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	err := untar(ctx, file, configFilesRestoration, filesRestoration, dbRestoration)
	if err != nil {
		fmt.Println(err)
		return
	}

	// remove decrypted file
	if isEncrypted {
		err = os.Remove(file)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Printf("\n %s Done\n", color.GreenString("✓"))
}

func decrypt(file string, key *string) (string, error) {
	encryptedExtension := file[len(file)-4:]

	if encryptedExtension != ".enc" || key == nil || *key == "" {
		return file, nil
	}

	// decrypt in an hidden file
	outputFile := "." + file[:len(file)-4]
	command := fmt.Sprintf("openssl enc -d -aes-256-cbc -salt -k %s -in %s -out %s", *key, file, outputFile)
	cmd := exec.Command("bash", "-c", command)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		removeErr := os.Remove(outputFile)
		if removeErr != nil {
			fmt.Println(removeErr)
		}

		return "", fmt.Errorf(fmt.Sprint(err) + ": " + stderr.String())
	}
	fmt.Printf("\n %s %s decrypted\n", color.GreenString("✓"), file)

	return outputFile, nil
}

func untar(ctx domain.ExecutionContext, tarball string, configFilesRestoration bool, filesRestoration bool, dbRestoration bool) error {
	// open the tarball
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()

	// gunzip
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// read the tarball
	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		info := header.FileInfo()

		// config
		if configFilesRestoration {
			if strings.HasPrefix(header.Name, "config/") {
				dest := strings.Replace(header.Name, "config/", "", 1)

				fmt.Printf(" → Restoring %s\n", dest)

				err := copyFile(dest, tarReader, info)
				if err != nil {
					return err
				}
			}
		}

		// files
		if filesRestoration {
			if strings.HasPrefix(header.Name, "files/") {
				dest := strings.Replace(header.Name, "files/", "", 1)

				fmt.Printf(" → Restoring %s\n", dest)

				err := copyFile(dest, tarReader, info)
				if err != nil {
					return err
				}
			}
		}

		// databases
		if dbRestoration {
			if strings.HasPrefix(header.Name, "databases/") && !info.IsDir() {
				dumpPath := strings.Replace(header.Name, "databases/", "", 1)
				// separate path components to get dump info
				comps := strings.Split(dumpPath, string(filepath.Separator))

				for _, dbBackup := range config.Get().BackupConfig.Databases {
					// search the container name
					if comps[0] == dbBackup.Container {

						fmt.Printf("\n → Restoring %s\n", dbBackup.Container)

						// get container id
						containerID, err := utils.GetContainerID(dbBackup.Container, ctx)
						if err != nil {
							fmt.Println("Unable to get the 'db' container id")
							continue
						}

						if strings.Contains(comps[1], "mongo") {
							restoreMongo(containerID, tarReader)
						} else if strings.Contains(comps[1], ".dump") {
							restorePostgres(ctx, containerID, comps[1], tarReader)
						} else if strings.Contains(comps[1], "sql") {
							// comps[1] is the filename of the dump (containing the database name, e.g. db.sql)
							restoreMySQL(ctx, containerID, comps[1], tarReader)
						} else {
							fmt.Println("Unrecognized db backup.")
						}

					}
				}
			}
		}

	}

	return nil
}

func copyFile(dest string, source io.Reader, sourceInfo os.FileInfo) error {
	if sourceInfo.IsDir() {
		return nil
	}
	dir := dest
	if !sourceInfo.IsDir() {
		dir = filepath.Dir(dest)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, sourceInfo.Mode())
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, source)
	if err != nil {
		return err
	}

	return nil
}

func restoreMongo(containerID string, mongoArchiveReader *tar.Reader) {
	cmd := domain.NewCommand([]string{"docker", "exec", "-i", containerID, "mongorestore", "--archive", "--gzip"})
	cmd.ExecuteWithStdin(mongoArchiveReader)
}

func restoreMySQL(ctx domain.ExecutionContext, containerID string, dumpFilename string, mysqlDumpReader *tar.Reader) {

	containerConfig, err := utils.GetContainerConfig(containerID, ctx)
	if err != nil {
		return
	}

	password := ""
	if value, ok := containerConfig.Env["MYSQL_ROOT_PASSWORD"]; ok {
		password = value
	}

	ext := filepath.Ext(dumpFilename)
	database := strings.Replace(dumpFilename, ext, "", 1)

	// backward compatibility, supporting previous filename (dump.sql)
	if database == "dump" {
		database = "db"
	}

	cmd := domain.NewCommand([]string{"docker", "exec", "-i", containerID, "mysql", fmt.Sprintf("--password=%s", password), database})
	cmd.ExecuteWithStdin(mysqlDumpReader)
}

func restorePostgres(ctx domain.ExecutionContext, containerID string, dumpFilename string, postgresDumpReader *tar.Reader) {

	containerConfig, err := utils.GetContainerConfig(containerID, ctx)
	if err != nil {
		return
	}

	user := "postgres"
	if value, ok := containerConfig.Env["POSTGRES_USER"]; ok {
		user = value
	}

	password := ""
	if value, ok := containerConfig.Env["POSTGRES_PASSWORD"]; ok {
		password = value
	}

	ext := filepath.Ext(dumpFilename)
	database := strings.Replace(dumpFilename, ext, "", 1)

	cmd := domain.NewCommand([]string{"docker", "exec", "-i", "-e", fmt.Sprintf("PGPASSWORD=\"%s\"", password), containerID, "pg_restore", fmt.Sprintf("--username=%s", user), "-d", database, "-c"})
	cmd.ExecuteWithStdin(postgresDumpReader)
}
