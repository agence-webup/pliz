package actions

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Songmu/prompter"

	"webup/pliz/config"
	"webup/pliz/domain"
	"webup/pliz/utils"
)

// RestoreActionHandler handle the action for 'pliz restore'
func RestoreActionHandler(ctx domain.ExecutionContext, file string) {

	if ctx.IsProd() {
		ok := prompter.YN("You're in production. Are you sure you want to continue?", false)
		if !ok {
			return
		}
	}

	err := untar(ctx, file)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("\n ✓ Done")
}

func ungzipReader(source string) (*gzip.Reader, error) {
	reader, err := os.Open(source)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}

	return archive, nil
}

func untar(ctx domain.ExecutionContext, tarball string) error {

	// choices
	fmt.Println(" ▶ ️ Choose what you want to restore:")
	configFilesRestoration := prompter.YN("     - configuration files", true)
	filesRestoration := prompter.YN("     - others files", true)
	dbRestoration := prompter.YN("     - database dumps", true)

	fmt.Printf("\n\n")

	reader, err := ungzipReader(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()

	tarReader := tar.NewReader(reader)

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
			if strings.HasPrefix(header.Name, "databases/") {
				dumpPath := strings.Replace(header.Name, "databases/", "", 1)
				// separate path components to get dump info
				comps := strings.Split(dumpPath, string(filepath.Separator))

				for _, dbBackup := range config.Get().BackupConfig.Databases {
					// search the container name
					if comps[0] == dbBackup.Container {

						fmt.Printf("\n → Restoring %s\n", dbBackup.Container)

						// get container id
						cmd := domain.NewComposeCommand([]string{"ps", "-q", dbBackup.Container}, ctx.IsProd())
						containerID, err := cmd.GetResult()
						if err != nil {
							fmt.Println("Unable to get the 'db' container id")
						}

						if strings.Contains(comps[1], "mongo") {
							restoreMongo(containerID, tarReader)
						} else if strings.Contains(comps[1], "dump") {
							restoreMySQL(ctx, containerID, tarReader)
						} else {
							fmt.Println("Unrecognized db backup.")
						}

						continue
					}
				}
			}
		}

	}

	return nil
}

func copyFile(dest string, source io.Reader, sourceInfo os.FileInfo) error {

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

func restoreMySQL(ctx domain.ExecutionContext, containerID string, mysqlDumpReader *tar.Reader) {

	containerConfig := utils.GetContainerConfig(containerID, ctx)

	password := ""
	if value, ok := containerConfig.Env["MYSQL_ROOT_PASSWORD"]; ok {
		password = value
	}

	database := "db"
	if value, ok := containerConfig.Env["MYSQL_DATABASE"]; ok {
		database = value
	}

	cmd := domain.NewCommand([]string{"docker", "exec", "-i", containerID, "mysql", fmt.Sprintf("--password=%s", password), database})
	cmd.ExecuteWithStdin(mysqlDumpReader)
}