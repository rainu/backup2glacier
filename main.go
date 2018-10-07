package main

import (
	"backup2glacier/cli"
	"backup2glacier/config"
)

func main() {
	cfg := config.NewConfig()

	var cliAction cli.CliAction

	switch cfg.General.Action {
	case config.ActionCreate:
		cliAction = cli.NewCreateAction()
	case config.ActionShow:
		cliAction = cli.NewShowAction()
	case config.ActionList:
		cliAction = cli.NewListAction()
	default:
		panic("This should never happen!")
	}

	cliAction.Validate(cfg)
	cliAction.Do(cfg)
}
