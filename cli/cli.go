package cli

import (
	"errors"
	"fmt"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/latam-airlines/crane/cluster"
	"github.com/latam-airlines/crane/util"
	"github.com/latam-airlines/crane/version"
	"github.com/codegangsta/cli"
	_ "github.com/latam-airlines/mesos-framework-factory/marathon"
	"github.com/latam-airlines/mesos-framework-factory/factory"
)

var stackManager *cluster.StackManager
var logFile *os.File = nil

type logConfig struct {
	LogLevel     string
	LogFormatter string
	LogColored   bool
	LogOutput    string
}

func dockerCfgPath() string {
	p := path.Join(os.Getenv("HOME"), ".docker", "config.json")
	if err := util.FileExists(p); err != nil {
		p = path.Join(os.Getenv("HOME"), ".dockercfg")
	}

	return p
}

func setupLogger(debug bool, config logConfig) error {
	var err error

	if util.Log.Level, err = log.ParseLevel(config.LogLevel); err != nil {
		return err
	}

	if debug {
		util.Log.Level = log.DebugLevel
	}

	switch config.LogFormatter {
	case "text":
		formatter := new(log.TextFormatter)
		formatter.ForceColors = config.LogColored
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

	switch config.LogOutput {
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

func globalFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Modo de verbosidad debug",
		},
		cli.StringSliceFlag{
			Name:  "endpoint, ep",
			Usage: "Endpoint de la API del Scheduler",
		},
		cli.StringFlag{
			Name:  "framework",
			Usage: "Scheduler you want to use to orchestrate your containers",
		},
		cli.BoolFlag{
			Name:  "tls",
			Usage: "Utiliza TLS en la comunicacion con los Endpoints",
		},
		cli.BoolFlag{
			Name:   "tlsverify",
			Usage:  "Utiliza TLS Verify en la comunicacion con los Endpoints",
			EnvVar: "DOCKER_TLS_VERIFY",
		},
		cli.StringFlag{
			Name:   "cert_path",
			Usage:  "Directorio con los certificados",
			EnvVar: "DOCKER_CERT_PATH",
		},
		cli.StringFlag{
			Name:   "tlscacert",
			Value:  "ca.pem",
			Usage:  "Ruta relativa del archivo con el certificado CA",
			EnvVar: "DEPLOYER_CERT_CA",
		},
		cli.StringFlag{
			Name:   "tlscert",
			Value:  "cert.pem",
			Usage:  "Ruta relativa del arhivo con el certificado cliente",
			EnvVar: "DEPLOYER_CERT_CERT",
		},
		cli.StringFlag{
			Name:   "tlskey",
			Value:  "key.pem",
			Usage:  "Ruta relativa del arhivo con la llave del certificado cliente",
			EnvVar: "DEPLOYER_CERT_KEY",
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
		cli.IntFlag{
			Name:   "deploy-timeout",
			Value:  30,
			Usage:  "Deploy timeout in seconds, default 30 seconds",
			EnvVar: "DEPLOYER_TIMEOUT",
		},
	}

	return flags
}

func buildCertPath(certPath string, file string) string {
	if file == "" {
		return ""
	}

	if certPath != "" {
		return certPath + "/" + file
	}

	return file
}

func setupGlobalFlags(c *cli.Context) error {
	var config logConfig = logConfig{}
	config.LogLevel = c.String("log-level")
	config.LogFormatter = c.String("log-formatter")
	config.LogColored = c.Bool("log-colored")
	config.LogOutput = c.String("log-output")

	var err error

	if err = setupLogger(c.Bool("debug"), config); err != nil {
		fmt.Println("Nivel de log invalido")
		return err
	}

	frameworkType := c.String("framework")

	stackManager = cluster.NewStackManager()

	for _, ep := range c.StringSlice("endpoint") {
		util.Log.Infof("Configurando el endpoint de Docker %s", ep)
		params := make(map[string]interface{})
		params["address"] = ep
		params["deploy-timeout"] = c.Int("deploy-timeout")
		clusterFramework, err := factory.Create(frameworkType, params)
		if err != nil {
			return errors.New(fmt.Sprintf("Error creating framework %s in %s. %s", frameworkType, ep, err.Error()))
		}
		stackManager.AppendStack(clusterFramework)
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
		return setupGlobalFlags(c)
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
		fmt.Println(err)
		util.Log.Fatalln(err)
	}
}
