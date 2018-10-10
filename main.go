package main

import (
	"backup2glacier/cli"
	"backup2glacier/config"
	"os"
)

func main() {
	os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	os.Setenv("AWS_PROFILE", "default")

	cfg := config.NewConfig()

	var cliAction cli.CliAction

	switch cfg.Action {
	case config.ActionCreate:
		cliAction = cli.NewCreateAction()
	case config.ActionGet:
		cliAction = cli.NewGetAction()
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
