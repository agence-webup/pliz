package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"webup/pliz/config"
	"webup/pliz/domain"
)

type containerConfig struct {
	Env   []string
	Image string
}

func BackupActionHandler(ctx domain.ExecutionContext) func() {
	return func() {

		// dump
		makeDump(ctx)

	}
}

func makeDump(ctx domain.ExecutionContext) {
	dbContainer := config.Get().Containers.Db

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
		fmt.Println("Unable to get the 'db' container id")
	}

	// parse the json
	var config containerConfig
	json.NewDecoder(strings.NewReader(configJson)).Decode(&config)

	// check for env variable to connect to DB
	password := ""
	database := dbContainer
	for _, data := range config.Env {
		items := strings.SplitN(data, "=", 2)
		if items[0] == "MYSQL_ROOT_PASSWORD" {
			password = items[1]
		} else if items[0] == "MYSQL_DATABASE" {
			database = items[1]
		}
	}

	// get associated network
	// NOTE: assumes that the container is associated to a single network
	cmd = domain.NewCommand([]string{"docker", "inspect", "--format", "{{range $net, $data := .NetworkSettings.Networks}}{{$net}}{{end}}", containerId})
	network, err := cmd.GetResult()
	if err != nil {
		fmt.Println("Unable to get the associated network")
	}

	// cmd = domain.NewContainerCommand(dbContainer, []string{"mysqldump", "-h", "db", "--password=truite", "db"}, ctx.IsProd())
	cmd = domain.NewCommand([]string{"docker", "run", "--rm", "--net", network, config.Image, "mysqldump", "-h", dbContainer, fmt.Sprintf("--password=%s", password), database})

	file, err := ioutil.TempFile("", "plizdump")
	if err != nil {
		fmt.Println("Unable to create a tmp file:")
		fmt.Println(err)
	}
	defer file.Close()

	if err := cmd.WriteResultToFile(file); err != nil {
		os.Remove(file.Name())
		fmt.Println(err)
		return
	}

	os.Rename(file.Name(), "dump.sql")
}
