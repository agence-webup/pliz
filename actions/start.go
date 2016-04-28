package actions

import (
	"webup/pliz/config"
	"webup/pliz/domain"
)

func StartActionHandler(prod bool) {
	cmd := domain.NewComposeCommand([]string{"up", "-d", config.Get().StartupContainer}, prod)
	cmd.Execute()
}
