package cli

import (
	"io/ioutil"
	"os"
	"path/filepath"

	valid "github.com/asaskevich/govalidator"
	"github.com/codegangsta/cli"
	"github.com/latam-airlines/crane/cluster"
	"github.com/latam-airlines/crane/configuration"
	"github.com/latam-airlines/crane/logger"
	"github.com/latam-airlines/crane/version"
	"gopkg.in/yaml.v2"
)

var stackManager cluster.CraneManager

type parseConfig func(configFile string) (*configuration.Configuration, error)

func readConfiguration(configFile string) (*configuration.Configuration, error) {
	_, err := os.Stat(configFile)
	if os.IsNotExist(err) {
		return nil, err
	}

	configFile, err = filepath.Abs(configFile)
	if err != nil {
		return nil, err
	}

	var yamlFile []byte
	if yamlFile, err = ioutil.ReadFile(configFile); err != nil {
		return nil, err
	}

	var config configuration.Configuration
	if err = yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, err
	}

	if _, err := valid.ValidateStruct(config); err != nil {
		return nil, err
	}

	return &config, nil
}

func globalFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Mode of debug verbosity",
		},
		cli.StringFlag{
			Name:  "config",
			Value: "crane.yml",
			Usage: "Path to config-file",
		},
	}

	return flags
}

func setupApplication(c *cli.Context, parser parseConfig) error {
	var appConfig *configuration.Configuration
	var err error
	if appConfig, err = parser(c.String("config")); err != nil {
		return err
	}

	if err := logger.Configure(appConfig.Logging, c.Bool("debug")); err != nil {
		return err
	}

	stackManager, err = cluster.NewStackManager(appConfig)
	if err != nil {
		return err
	}
	return nil
}

func RunApp() {

	app := cli.NewApp()
	app.Name = "cloud-crane"
	app.Usage = "Multi-Scheduler Orchestrator"
	app.Version = version.VERSION + " (" + version.GITCOMMIT + ")"

	app.Flags = globalFlags()

	app.Before = func(c *cli.Context) error {
		return setupApplication(c, readConfiguration)
	}

	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		logger.Instance().Fatalln(err)
	}
}
