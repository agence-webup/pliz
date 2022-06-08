package actions

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path"
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

	dpath, dfile := path.Split(file)
	isEncrypted := false
	encryptedFile := file
	decryptedFile := ""
	if len(dfile) > 4 {
		encryptedExtension := dfile[len(dfile)-4:]
		if encryptedExtension != ".enc" && key != nil && *key != "" {
			fmt.Printf(" %s This is not a .enc file, skip deciphering\n", color.RedString("✗"))
		}
		isEncrypted = encryptedExtension == ".enc" && key != nil && *key != ""
		decryptedFile = dpath + "." + dfile[:len(dfile)-4]
	}

	// decrypt in an hidden file
	if isEncrypted {
		err := decrypt(encryptedFile, decryptedFile, key)
		if err != nil {
			fmt.Println(err)
			return
		}

		file = decryptedFile
	}

	err := untar(ctx, file, configFilesRestoration, filesRestoration, dbRestoration)
	if err != nil {
		fmt.Println(err)
		return
	}

	// remove decrypted file
	if isEncrypted {
		err = os.Remove(decryptedFile)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Printf("\n %s Done\n", color.GreenString("✓"))
}

func decrypt(encryptedFile string, decryptedFile string, key *string) error {
	infile, err := os.Open(encryptedFile)
	if err != nil {
		log.Fatal(err)
	}
	outfile, err := os.OpenFile(decryptedFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}

	aesKey := []byte(*key)
	hmacKey := aesKey
	err = utils.Decrypt(infile, outfile, aesKey, hmacKey)
	infile.Close()
	outfile.Close()

	if err != nil {
		removeDecryptedFile(decryptedFile)
		return fmt.Errorf("Unable to decrypt the file %s\n%s\n", decryptedFile, err)
	}
	fmt.Printf("\n %s %s decrypted\n", color.GreenString("✓"), encryptedFile)

	return nil
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

func removeDecryptedFile(file string) {
	if _, err := os.Stat(file); err == nil {
		removeErr := os.Remove(file)
		if removeErr != nil {
			fmt.Println(removeErr)
		}
	}
}
