package cli

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	valid "github.com/asaskevich/govalidator"
	"github.com/codegangsta/cli"
	"github.com/latam-airlines/crane/cluster"
	"github.com/latam-airlines/crane/configuration"
	"github.com/latam-airlines/crane/util"
	"github.com/latam-airlines/crane/version"
	"gopkg.in/yaml.v2"
)

var stackManager cluster.CraneManager
var logFile *os.File

func setupLogger(config configuration.Loggging, debug bool) error {
	var err error

	if util.Log.Level, err = log.ParseLevel(config.Level); err != nil {
		return err
	}

	if debug {
		util.Log.Level = log.DebugLevel
	}

	switch config.Formatter {
	case "text":
		formatter := new(log.TextFormatter)
		formatter.ForceColors = config.Colored
		formatter.FullTimestamp = true
		util.Log.Formatter = formatter
		break
	case "json":
		formatter := new(log.JSONFormatter)
		util.Log.Formatter = formatter
		break
	default:
		return errors.New("Formato de lo log desconocido")
	}

	switch config.Output {
	case "console":
		util.Log.Out = os.Stdout
		break
	case "file":
		util.Log.Out = logFile
		break
	default:
		return errors.New("Output de logs desconocido")
	}

	return nil
}

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

	if err := setupLogger(appConfig.Logging, c.Bool("debug")); err != nil {
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
		err := setupApplication(c, readConfiguration)
		if err != nil {
			util.Log.Fatalln(err)
		}
		return nil
	}

	app.Commands = commands

	var err error
	logFile, err = os.OpenFile("cloud-crane.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		util.Log.Warnln("Error al abrir el archivo")
	} else {
		defer logFile.Close()
	}

	err = app.Run(os.Args)
	if err != nil {
		util.Log.Fatalln(err)

	}
}
