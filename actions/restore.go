package actions

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"webup/pliz/config"
	"webup/pliz/domain"
)

func RestoreActionHandler(file string) {
	err := untar(file)
	fmt.Println(err)
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

func untar(tarball string) error {

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
		if strings.HasPrefix(header.Name, "config/") {
			dest := strings.Replace(header.Name, "config/", "", 1)
			err := copyFile(dest, tarReader, info)
			if err != nil {
				return err
			}
		}

		// files
		if strings.HasPrefix(header.Name, "files/") {
			dest := strings.Replace(header.Name, "files/", "", 1)
			err := copyFile(dest, tarReader, info)
			if err != nil {
				return err
			}
		}

		// databases
		if strings.HasPrefix(header.Name, "databases/") {
			dumpPath := strings.Replace(header.Name, "databases/", "", 1)
			// separate path components to get dump info
			comps := strings.Split(dumpPath, string(filepath.Separator))

			for _, dbBackup := range config.Get().BackupConfig.Databases {
				// search the container name
				if comps[0] == dbBackup.Container {
					if strings.Contains(comps[1], "mongo") {
						restoreMongo(dbBackup, tarReader)
					}
					continue
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

func restoreMongo(dbBackup domain.DatabaseBackupConfig, tarReader *tar.Reader) {
	cmd := domain.NewComposeCommand([]string{"ps", "-q", dbBackup.Container}, false) // false => isProd
	containerID, err := cmd.GetResult()
	if err != nil {
		fmt.Println("Unable to get the 'db' container id")
	}

	cmd = domain.NewCommand([]string{"docker", "exec", "-i", containerID, "mongorestore", "--archive", "--gzip"})
	rawCmd := cmd.GetRawExecCommand()
	rawCmd.Stderr = os.Stderr
	rawCmd.Stdout = os.Stdout

	stdinReader := bufio.NewReader(tarReader)
	stdinWriter, err := rawCmd.StdinPipe()
	if err != nil {
		fmt.Println(err)
	}

	_, err = stdinReader.WriteTo(stdinWriter)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Executing: %s\n", cmd)
	rawCmd.Start()

	// close writer to indicate that stdin is finished (avoiding hanging of the exec cmd)
	stdinWriter.Close()

	rawCmd.Wait()

}
