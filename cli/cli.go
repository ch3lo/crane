package cli

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/latam-airlines/crane/cluster"
	"github.com/latam-airlines/crane/configuration"
	"github.com/latam-airlines/crane/util"
	"github.com/latam-airlines/crane/version"
)

var stackManager cluster.CraneManager
var logFile *os.File

type logConfig struct {
	level     string
	Formatter string
	colored   bool
	output    string
	debug     bool
}

func setupLogger(config logConfig) error {
	var err error

	if util.Log.Level, err = log.ParseLevel(config.level); err != nil {
		return err
	}

	if config.debug {
		util.Log.Level = log.DebugLevel
	}

	switch config.Formatter {
	case "text":
		formatter := new(log.TextFormatter)
		formatter.ForceColors = config.colored
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

	switch config.output {
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

	return &config, nil
}

func globalFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Modo de verbosidad debug",
		},
		cli.StringFlag{
			Name:  "config",
			Value: "crane.yml",
			Usage: "Ruta del archivo de configuración",
		},
		cli.StringFlag{
			Name:   "log-level",
			Value:  "info",
			Usage:  "Nivel de verbosidad de log",
			EnvVar: "DEPLOYER_LOG_LEVEL",
		},
		cli.StringFlag{
			Name:   "log-formatter",
			Value:  "text",
			Usage:  "Formato de log",
			EnvVar: "DEPLOYER_LOG_FORMATTER",
		},
		cli.BoolFlag{
			Name:   "log-colored",
			Usage:  "Coloreo de log :D",
			EnvVar: "DEPLOYER_LOG_COLORED",
		},
		cli.StringFlag{
			Name:   "log-output",
			Value:  "file",
			Usage:  "Output de los logs. console | file",
			EnvVar: "DEPLOYER_LOG_OUTPUT",
		},
	}

	return flags
}

func setupApplication(c *cli.Context, parser parseConfig) error {
	logConfig := logConfig{}
	logConfig.level = c.String("log-level")
	logConfig.Formatter = c.String("log-formatter")
	logConfig.colored = c.Bool("log-colored")
	logConfig.output = c.String("log-output")
	logConfig.debug = c.Bool("debug")

	err := setupLogger(logConfig)
	if err != nil {
		return err
	}

	var appConfig *configuration.Configuration
	if appConfig, err = parser(c.String("config")); err != nil {
		return err
	}

	stackManager = cluster.NewStackManager(appConfig)
	return nil
}

// RunApp Entrypoint de la Aplicacion.
// Procesa los comandos y sus argumentos
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
