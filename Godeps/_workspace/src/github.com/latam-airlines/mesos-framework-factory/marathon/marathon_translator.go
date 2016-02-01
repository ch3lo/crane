package marathon

import (
	"github.com/latam-airlines/crane/Godeps/_workspace/src/github.com/gambol99/go-marathon"
	"github.com/latam-airlines/crane/Godeps/_workspace/src/github.com/latam-airlines/mesos-framework-factory"
	"github.com/latam-airlines/crane/Godeps/_workspace/src/github.com/latam-airlines/mesos-framework-factory/utils"
	"strconv"
	"strings"
)

func translateServiceConfig(config *framework.ServiceConfig, instances int) *marathon.Application {
	application := marathon.NewDockerApplication()
	imageWithTag := config.ImageName + ":" + config.Tag
	labels := map[string]string{
		"image_name": config.ImageName,
		"image_tag":  config.Tag,
	}

	application.Name(config.ServiceID)

	if config.Constraints["slave_name"] != "" {
		labels["slave_name"] = config.Constraints["slave_name"]
	}

	// default value for CPU
	if config.CPUShares != 0.0 {
		application.CPU(config.CPUShares)
	} else {
		application.CPU(0.25)
	}

	if config.DockerCfg != "" {
		//uris := make([]string, 1)
		//uris = append(uris, config.DockerCfg)
		//application.Uris = []string{config.DockerCfg}
		application.Uris = append(application.Uris, config.DockerCfg)
	}
	application.Memory(float64(config.Memory))
	application.Count(instances)
	application.Env = utils.StringSlice2Map(config.Envs)
	application.Env["SERVICE_NAME"] = config.ServiceID
	application.Labels = labels

	//upgradeStrategy
	application.UpgradeStrategy = &marathon.UpgradeStrategy{
		MinimumHealthCapacity: config.MinimumHealthCapacity,
		MaximumOverCapacity:   config.MaximumOverCapacity,
	}

	//application.RequirePorts = true
	populateConstraints(application, config)
	// add the docker container
	application.Container.Docker.Container(imageWithTag)
	application.Container.Docker.PortMappings = createPorMappings(config.Publish)
	application.Container.Docker.Parameters = populateParameters(config)
	return application
}

func populateParameters(cfg *framework.ServiceConfig) []*marathon.Parameters {
	//in future we will add more params for sure
	return getSyslogParams(cfg)
}

func getSyslogParams(cfg *framework.ServiceConfig) []*marathon.Parameters {
	syslogConfigurer := utils.CreateSyslogConfigurer(cfg)
	params := make([]*marathon.Parameters, 0)
	for _, syslogParam := range syslogConfigurer.GetFlags() {
		params = append(params, &marathon.Parameters{Key: syslogParam.Key, Value: syslogParam.Value})
	}
	return params
}

func populateConstraints(app *marathon.Application, cfg *framework.ServiceConfig) {
	if cfg.Constraints != nil && len(cfg.Constraints) > 0 {
		constraints := make([][]string, len(cfg.Constraints))
		for i := range constraints {
			constraints[i] = make([]string, 3)
		}

		idx := 0
		for key, val := range cfg.Constraints {
			constraints[idx][0] = key
			constraints[idx][1] = "CLUSTER"
			constraints[idx][2] = val
			idx++
		}
		app.Constraints = constraints
	}
}

func createPorMappings(ports []string) []*marathon.PortMapping {
	if ports == nil || len(ports) == 0 {
		return nil
	}

	portMappings := make([]*marathon.PortMapping, len(ports))
	for i, val := range ports {
		portConfig := strings.Split(val, "/")
		iPort, _ := strconv.Atoi(portConfig[0])
		portMappings[i] = createPortMapping(iPort, portConfig[1])
	}

	return portMappings
}

func createPortMapping(containerPort int, protocol string) *marathon.PortMapping {
	return &marathon.PortMapping{
		ContainerPort: containerPort,
		HostPort:      0,
		ServicePort:   0,
		Protocol:      protocol,
	}
}
