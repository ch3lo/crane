package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"strconv"
	"regexp"

	"github.com/latam-airlines/crane/cluster"
	"github.com/latam-airlines/crane/util"
	"github.com/codegangsta/cli"
	"github.com/latam-airlines/mesos-framework-factory"
)

func handleDeploySigTerm(sm *cluster.StackManager) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		sm.Rollback()
		os.Exit(1)
	}()
}

func deployFlags() []cli.Flag {
	return []cli.Flag{
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
		cli.IntFlag{
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
	}
}

func deployBefore(c *cli.Context) error {
	if c.String("image") == "" {
		return errors.New("El nombre de la imagen esta vacio")
	}

	if c.String("tag") == "" {
		return errors.New("El TAG de la imagen esta vacio")
	}
	
	if c.String("memory") != "" {
		if _, err := strconv.ParseInt(c.String("memory"), 10, 64); err != nil {
			return errors.New("Valor del parámetro memory invalido")
		}
	}

	for _, file := range c.StringSlice("env-file") {
		if err := util.FileExists(file); err != nil {
			return errors.New(fmt.Sprintf("El archivo %s con variables de entorno no existe", file))
		}
	}

	return nil
}

type callbackResume struct {
	RegisterId string `json:"RegisterId"`
	Address    string `json:"Address"`
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
			cfg.Publish[i]=port
		} else {
			return errors.New("Port does not match format, ie. 8080/tcp")
		}
        }
	return nil
}

func deployCmd(c *cli.Context) {

	envs, err := util.ParseMultiFileLinesToArray(c.StringSlice("env-file"))
	if err != nil {
		util.Log.Fatalln("No se pudo procesar el archivo con variables de entorno", err)
	}

	for _, v := range c.StringSlice("env") {
		envs = append(envs, v)
	}

	serviceConfig := framework.ServiceConfig{
		CPUShares: c.Int("cpu"),
		Envs:      envs,
		ImageName: c.String("image"),
		Tag:       c.String("tag"),
	}
	serviceConfig.ConvertImageTagToServiceId()
	applyPorts(c.StringSlice("port"), &serviceConfig)
	if c.String("memory") != "" {
		n, _ := strconv.ParseInt(c.String("memory"), 10, 64)
		serviceConfig.Memory = int64(n)
	}

	util.Log.Debugf("La configuración del servicio es: %#v", serviceConfig.String())

	handleDeploySigTerm(stackManager)
	if stackManager.Deploy(serviceConfig, c.Int("instances"), c.Float64("tolerance")) {
		services := stackManager.DeployedContainers()
		var resume []callbackResume
		for _, service := range services {
			for _, instance := range service.Instances {
				for _, val := range instance.Ports {
					util.Log.Infof("Se desplegó %s en host %s y dirección %s", instance.ID, instance.Host, val)
					instanceInfo := callbackResume{
						RegisterId: instance.ID,
						Address:    string(val.Internal),
					}
					resume = append(resume, instanceInfo)
				}
			}
		}
		jsonResume, _ := json.Marshal(resume)

		fmt.Println(string(jsonResume))
	} else {
		util.Log.Fatalln("Proceso de deploy con errores")
	}
}
