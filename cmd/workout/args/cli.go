package args

import (
	"fmt"

	"git.rpjosh.de/RPJosh/RPdb/v4/go/pkg/cli"
	"git.rpjosh.de/RPJosh/workout/internal/api"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/dbutils"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/tests"
)

type Cli struct {

	// Sub commands
	User *User `cli:"user,u"`

	// If the program is called in auto-completion mode
	AutoComplete bool

	// App configuration
	Config *models.AppConfig
}

func (cli *Cli) Help() string {
	return (`
Syntax: ProgramName user\|anything [options]

To get a help to the various options, execute these again with the parameter --help.
For example: ProgramName user --help

  user      u     |Create and manage users
	`)
}

func (cli *Cli) EnableAutoComplete() {
	cli.AutoComplete = true
}

// InjectApi injects all fields for the struct type
// [router.ApiRequestler] with a mocked one to also
// use API endpoints in CLI mode
func (cli *Cli) InjectApi(dst router.ApiRequestler) {
	a := api.Api{Config: cli.Config}

	conf := &tests.RouterConfig{
		Db: dbutils.New(a.GetDb()).Db,
		User: models.WebUser{
			User: &models.User{
				Id:   1,
				Name: "HI",
			},
		},
	}

	tests.InjectRequestDataWithConfig(dst, conf)
}

func ParseArgs(config *models.AppConfig, args []string) error {
	cl := &Cli{
		User:   &User{},
		Config: config,
	}

	if cli.ParseParams(args, cl) < 0 {
		return fmt.Errorf("")
	}

	return nil
}
