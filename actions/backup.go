package actions

import (
	"fmt"
	"webup/pliz/config"
	"webup/pliz/domain"
)

func BackupActionHandler(ctx domain.ExecutionContext) func() {
	return func() {

		dbContainer := config.Get().Containers.Db

		// fetch the container id for db
		cmd := domain.NewComposeCommand([]string{"ps", "-q", dbContainer}, ctx.IsProd())
		containerId, err := cmd.GetResult()
		if err != nil {
			fmt.Println("Unable to get the 'db' container id")
		}

		fmt.Println(containerId)

		cmd = domain.NewContainerCommand(dbContainer, []string{"mysqldump", "-h", "db", "--password=truite", "db"}, ctx.IsProd())
		cmd.Execute()
	}
}
