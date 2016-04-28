package actions

import (
	"webup/pliz/config"
	"webup/pliz/domain"
)

func StartActionHandler(prod bool, startAdditionalContainers bool) {

	args := []string{"up", "-d", config.Get().StartupContainer}

	if startAdditionalContainers {
		args = append(args, config.Get().AdditionalStartupContainers...)
	}

	cmd := domain.NewComposeCommand(args, prod)
	cmd.Execute()
}
