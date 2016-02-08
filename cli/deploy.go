package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/latam-airlines/crane/cluster"
	"github.com/latam-airlines/crane/logger"
	"github.com/latam-airlines/crane/util"
	"github.com/latam-airlines/mesos-framework-factory"
)

func handleDeploySigTerm(sm cluster.CraneManager) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		//sm.Rollback() XXX: To Fix This Rollback needs the current version of every Service
		os.Exit(1)
	}()
}

func deployFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "service-id",
			Usage: "Id of the service",
		},
		cli.StringFlag{
			Name:  "image",
			Usage: "Nombre de la imagen",
		},
		cli.StringFlag{
			Name:  "tag",
			Usage: "TAG de la imagen",
		},
		cli.StringSliceFlag{
			Name:  "port",
			Usage: "Puerto interno del contenedor a exponer en el Host",
		},
		cli.Float64Flag{
			Name:  "cpu",
			Value: 0,
			Usage: "Cantidad de CPU reservadas para el servicio.",
		},
		cli.StringFlag{
			Name:  "memory",
			Usage: "Cantidad de memoria principal (Unidades: M, m, MB, mb, GB, G) que puede utilizar el servicio. Mas info 'man docker-run' memory.",
		},
		cli.StringSliceFlag{
			Name:  "env-file",
			Usage: "Archivo con variables de entorno",
		},
		cli.StringSliceFlag{
			Name:  "env",
			Usage: "Variables de entorno en formato KEY=VALUE",
		},
		cli.IntFlag{
			Name:  "instances",
			Value: 1,
			Usage: "Total de servicios que se quieren obtener en cada uno de los stack.",
		},
		cli.Float64Flag{
			Name:  "tolerance",
			Value: 0.5,
			Usage: "Porcentaje de servicios que pueden fallar en el proceso de deploy por cada enpoint entregado." +
				"Este valor es respecto al total de instancias." +
				"Por ejemplo, si se despliegan 5 servicios y fallan ",
		},
		cli.StringSliceFlag{
			Name:  "constraint",
			Usage: "Add constraint to the deployment, ie --constraint=slave_name=beta4002, --constraint=hostname=UNIQUE",
		},
		cli.StringFlag{
			Name:  "beta",
			Usage: "Beta-Node to deploy to, ie --beta=beta4002",
		},
		cli.Float64Flag{
			Name:  "minimumHealthCapacity",
			Value: 1.0,
			Usage: "Number between 0 and 1 that is multiplied with the instance count. This is the minimum number of healthy nodes that do not sacrifice overall application purpose, ie --minimumHealthCapacity=0.8",
		},
		cli.Float64Flag{
			Name:  "maximumOverCapacity",
			Value: 0.2,
			Usage: "Number between 0 and 1 which is multiplied with the instance count. This is the maximum number of additional instances launched at any point of time during the upgrade process, i.e. maximumOverCapacity=0.2",
		},
		cli.StringFlag{
			Name:  "health-check-path",
			Usage: "path to the health check file. ex: --health-check-path=/v0/healthy",
		},
		cli.StringSliceFlag{
			Name:  "label",
			Usage: "Add the label to the deployment, ie --label=environment=beta, --label=test=beta",
		},
	}
}

func deployBefore(c *cli.Context) error {
	if c.String("service-id") == "" {
		return errors.New("Service-id is empty")
	}

	if c.String("image") == "" {
		return errors.New("The name of the image is empty")
	}

	if c.String("tag") == "" {
		return errors.New("The Tag of the image is empty")
	}

	if c.String("memory") != "" {
		if _, err := strconv.ParseInt(c.String("memory"), 10, 64); err != nil {
			return errors.New("Invalid value of paramter memory")
		}
	}

	if c.Float64("cpu") < 0 {
		return errors.New("Cpu flag value should not be negative")
	}

	if c.String("framework") == "marathon" && c.Float64("cpu") > 1.0 { // Fix this: framework flag does not exist anymore
		return errors.New("Cpu flag value should not be > 1.0 for marathon")
	} else if c.String("framework") == "swarm" && c.Float64("cpu") > 1024 { // Fix this: framework flag does not exist anymore
		return errors.New("Cpu flag value should not be > 1024.0 for swarm")
	}

	for _, file := range c.StringSlice("env-file") {
		if err := util.FileExists(file); err != nil {
			return errors.New(fmt.Sprintf("El archivo %s con variables de entorno no existe", file))
		}
	}

	if c.Float64("minimumHealthCapacity") < 0.0 || c.Float64("minimumHealthCapacity") > 1.0 {
		return errors.New("MinimumHealthCapacity flag value should be between 0.0 and 1.0")
	}

	if c.Float64("maximumOverCapacity") < 0.0 || c.Float64("maximumOverCapacity") > 1.0 {
		return errors.New("MaximumOverCapacity flag value should be between 0.0 and 1.0")
	}

	return nil
}

type callbackResume struct {
	Id      string `json:Id"`
	Address string `json:"Address"`
}

func applyPorts(ports []string, cfg *framework.ServiceConfig) error {
	if ports == nil || len(ports) == 0 {
		cfg.Publish = []string{"8080/tcp"}
		return nil
	}
	cfg.Publish = make([]string, len(ports))
	var validPort = regexp.MustCompile(`^[0-9]*\/(udp|tcp|UDP|TCP)$`)
	for i, port := range ports {
		if validPort.MatchString(port) {
			cfg.Publish[i] = port
		} else {
			return errors.New("Port does not match format, ie. 8080/tcp")
		}
	}
	return nil
}

func applyKeyValSliceFlag(sliceFlag []string, setAction func(newMap map[string]string)) error {
	configMap := make(map[string]string)

	for _, flag := range sliceFlag {
		if !strings.Contains(flag, "=") {
			return errors.New("The flag does not comply format key=value")
		}
		splits := strings.Split(flag, "=")
		configMap[splits[0]] = splits[1]
	}

	if setAction != nil {
		setAction(configMap)
	}
	return nil
}

func deployCmd(c *cli.Context) {

	envs, err := util.ParseMultiFileLinesToArray(c.StringSlice("env-file"))
	if err != nil {
		logger.Instance().Fatalln("No se pudo procesar el archivo con variables de entorno", err)
	}

	for _, v := range c.StringSlice("env") {
		envs = append(envs, v)
	}

	serviceConfig := framework.ServiceConfig{
		ServiceID: c.String("service-id"),
		CPUShares: c.Float64("cpu"),
		Envs:      envs,
		ImageName: c.String("image"),
		Tag:       c.String("tag"),
		MinimumHealthCapacity: c.Float64("minimumHealthCapacity"),
		MaximumOverCapacity:   c.Float64("maximumOverCapacity"),
		HealthCheckConfig:     &framework.HealthCheck{Path: c.String("health-check-path")},
	}
	applyPorts(c.StringSlice("port"), &serviceConfig)
	if c.String("memory") != "" {
		n, _ := strconv.ParseInt(c.String("memory"), 10, 64)
		serviceConfig.Memory = int64(n)
	}

	err = applyKeyValSliceFlag(c.StringSlice("constraint"), func(configMap map[string]string) {
		if configMap != nil && len(configMap) != 0 {
			serviceConfig.Constraints = configMap
		}
	})

	if err != nil {
		logger.Instance().Fatalln("Error reading constraints", err)
	}

	err = applyKeyValSliceFlag(c.StringSlice("label"), func(configMap map[string]string) {
		if configMap != nil && len(configMap) != 0 {
			serviceConfig.Labels = configMap
		}
	})

	if err != nil {
		logger.Instance().Fatalln("Error reading labels", err)
	}

	if c.String("beta") != "" {
		if serviceConfig.Labels == nil {
			serviceConfig.Labels = make(map[string]string)
		}
		if serviceConfig.Constraints == nil {
			serviceConfig.Constraints = make(map[string]string)
		}

		serviceConfig.Labels["slave_name"] = c.String("beta")
		serviceConfig.Constraints["slave_name"] = c.String("beta")
	}

	handleDeploySigTerm(stackManager)
	if stackManager.Deploy(serviceConfig, c.Int("instances"), c.Float64("tolerance")) {
		services := stackManager.DeployedContainers()
		var resume []callbackResume
		for _, service := range services {
			for _, instance := range service.Instances {
				for _, val := range instance.Ports {
					logger.Instance().Infof("Se desplegó %s en host %s y dirección %+v", instance.ID, instance.Host, val)
					instanceInfo := callbackResume{
						Id:      instance.ID,
						Address: instance.Host + ":" + strconv.FormatInt(val.Internal, 10),
					}
					resume = append(resume, instanceInfo)
				}
			}
		}
		jsonResume, _ := json.Marshal(resume)
		fmt.Println(string(jsonResume))
	} else {
		logger.Instance().Fatalln("Deployment-Process terminated with errors")
	}
}
