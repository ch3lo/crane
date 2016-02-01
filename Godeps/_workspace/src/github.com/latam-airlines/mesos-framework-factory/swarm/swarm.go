package swarm

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"regexp"

	"github.com/fsouza/go-dockerclient"
	"github.com/latam-airlines/mesos-framework-factory"
	"github.com/latam-airlines/mesos-framework-factory/factory"
	"github.com/latam-airlines/mesos-framework-factory/logger"
)

const frameworkID = "swarm"

func init() {
	factory.Register(frameworkID, &swarmCreator{})
}

// swarmCreator implementa la interfaz factory.FrameworkFactory
type swarmCreator struct{}

func (factory *swarmCreator) Create(parameters map[string]interface{}) (framework.Framework, error) {
	return NewFromParameters(parameters)
}

// parameters encapsula los parametros de configuracion de Swarm
type parameters struct {
	address   string
	authfile  string
	tlsverify bool
	tlscacert string
	tlscert   string
	tlskey    string
}

func dockerCfgPath() string {
	p := path.Join(os.Getenv("HOME"), ".docker", "config.json")

	if _, err := os.Stat(p); os.IsNotExist(err) {
		p = path.Join(os.Getenv("HOME"), ".dockercfg")
	}

	return p
}

// NewFromParameters construye un Scheduler a partir de un mapeo de parámetros
// Al menos se debe pasar como parametro address, ya que si no existe se retornara un error
// Si se pasa tlsverify como true los parametros tlscacert, tlscert y tlskey también deben existir
func NewFromParameters(params map[string]interface{}) (*Framework, error) {

	address, ok := params["address"]
	if !ok || fmt.Sprint(address) == "" {
		return nil, errors.New("Parametro address no existe")
	}

	authfile := dockerCfgPath()
	if af, ok := params["authfile"]; !ok || fmt.Sprint(af) == "" {
		logger.Instance().Warnln("Parametro authfile no existe o está vacio, utilizando su valor por defecto", authfile)
	} else {
		authfile = fmt.Sprint(af)
	}

	tlsverify := false
	if tlsv, ok := params["tlsverify"]; ok {
		tlsverify, ok = tlsv.(bool)
		if !ok {
			return nil, fmt.Errorf("El parametro tlsverify debe ser un boolean")
		}
	}

	var tlscacert interface{}
	var tlscert interface{}
	var tlskey interface{}

	if tlsverify {
		tlscacert, ok = params["tlscacert"]
		if !ok || fmt.Sprint(tlscacert) == "" {
			return nil, errors.New("Parametro tlscacert no existe")
		}

		tlscert, ok = params["tlscert"]
		if !ok || fmt.Sprint(tlscert) == "" {
			return nil, errors.New("Parametro tlscert no existe")
		}

		tlskey, ok = params["tlskey"]
		if !ok || fmt.Sprint(tlskey) == "" {
			return nil, errors.New("Parametro tlskey no existe")
		}
	}

	p := parameters{
		address:   fmt.Sprint(address),
		authfile:  authfile,
		tlsverify: tlsverify,
		tlscacert: fmt.Sprint(tlscacert),
		tlscert:   fmt.Sprint(tlscert),
		tlskey:    fmt.Sprint(tlskey),
	}

	return New(p)
}

func authConfig(authConfigPath string, registry string) (docker.AuthConfiguration, error) {
	var r io.Reader
	var err error

	logger.Instance().Infof("Obteniendo los parámetros de autenticación para el registro %s del archivo %s", registry, authConfigPath)
	if r, err = os.Open(authConfigPath); err != nil {
		return docker.AuthConfiguration{}, err
	}

	var authConfigs *docker.AuthConfigurations

	if authConfigs, err = docker.NewAuthConfigurations(r); err != nil {
		return docker.AuthConfiguration{}, err
	}

	for key := range authConfigs.Configs {
		if key == registry {
			return authConfigs.Configs[registry], nil
		}
	}

	return docker.AuthConfiguration{}, errors.New("No se encontraron las credenciales de autenticación")
}

// New instancia un nuevo cliente de Swarm
func New(params parameters) (*Framework, error) {
	swarm := &Framework{
		authConfigs: make(map[string]docker.AuthConfiguration),
	}

	var err error
	logger.Instance().Debugf("Configurando Swarm con los parametros %+v", params)
	if params.tlsverify {
		swarm.client, err = docker.NewTLSClient(params.address, params.tlscert, params.tlskey, params.tlscacert)
	} else {
		swarm.client, err = docker.NewClient(params.address)
	}
	if err != nil {
		return nil, err
	}

	registriesURL := []string{"https://registry.it.lan.com", "https://registry.dev.lan.com"}
	for _, v := range registriesURL {
		auth, err := authConfig(params.authfile, v)
		if err != nil {
			return nil, err
		}
		swarm.authConfigs[v] = auth
	}

	return swarm, nil
}

// Framework es una implementacion de framework.Framework
// Permite el la comunicación con la API de Swarm
type Framework struct {
	client      *docker.Client
	authConfigs map[string]docker.AuthConfiguration
}

// ID retorna el identificador del framework Swarm
func (s *Framework) ID() string {
	return frameworkID
}

func getStatus(s string) framework.InstanceStatus {
	upRegexp := regexp.MustCompile("^[u|U]p")
	status := framework.InstanceDown
	if upRegexp.MatchString(s) {
		status = framework.InstanceUp
	}
	return status
}

func getStatusFromState(state docker.State) framework.InstanceStatus {
	status := framework.InstanceDown
	if state.Running &&
		!state.Paused &&
		!state.Restarting {
		status = framework.InstanceUp
	}
	return status
}

func hostAndContainerName(fullName string) (string, string) {
	hostAndContainerName := regexp.MustCompile("^(?:/([\\w|_-]+))?/([\\w|_-]+)$")
	result := hostAndContainerName.FindStringSubmatch(fullName)
	host := "unknown"
	if result[1] != "" {
		host = result[1]
	}
	containerName := result[2]
	return host, containerName
}

func imageAndTag(fullImageName string) (string, string) {
	imageAndTagRegexp := regexp.MustCompile("^([\\w./_-]+)(?::([\\w._-]+))?$")
	result := imageAndTagRegexp.FindStringSubmatch(fullImageName)
	imageName := result[1]
	imageTag := "latest"
	if result[2] != "" {
		imageTag = result[2]
	}
	return imageName, imageTag
}

func mapPorts(apiPorts []docker.APIPort) map[string]framework.InstancePort {
	ports := make(map[string]framework.InstancePort)
	for _, v := range apiPorts {
		id := fmt.Sprintf("%d/%s", v.PrivatePort, v.Type)
		p, ok := ports[id]
		if !ok {
			ip := "127.0.0.1"
			if v.IP != "0.0.0.0" {
				ip = v.IP
			}
			p = framework.InstancePort{
				Advertise: ip,
				Internal:  v.PrivatePort,
				Type:      framework.NewInstancePortType(v.Type),
			}
		}
		p.Publics = append(p.Publics, v.PublicPort)
		ports[id] = p
	}
	return ports
}

func getService(id string, services []*framework.ServiceInformation) *framework.ServiceInformation {
	for key, srv := range services {
		if id == srv.ID {
			return services[key]
		}
	}
	return nil
}

func (s *Framework) FindServiceInformation(filter framework.ServiceInformationCriteria) ([]*framework.ServiceInformation, error) {
	logger.Instance().Debugln("Obteniendo el listado de contenedores")

	containers, err := s.client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return nil, err
	}

	var services []*framework.ServiceInformation
	for _, container := range containers {
		logger.Instance().Debugf("Filtrando el contenedor %+v", container)

		imageName, imageTag := imageAndTag(container.Image)
		sid := imageName + ":" + imageTag

		srv := getService(sid, services)
		if srv == nil {
			srv = &framework.ServiceInformation{
				ID:        sid,
				ImageName: imageName,
				ImageTag:  imageTag,
			}
			services = append(services, srv)
		}

		host, containerName := hostAndContainerName(container.Names[0])
		c := &framework.Instance{
			ID:            container.ID,
			Status:        getStatus(container.Status),
			Host:          host,
			ContainerName: containerName,
			Ports:         mapPorts(container.Ports),
		}
		logger.Instance().Debugf("Mapeando contenedor a %+v", c)
		srv.Instances = append(srv.Instances, c)
	}

	if filter == nil {
		return services, nil
	}

	filtered := filter.MeetCriteria(services)

	return filtered, nil
}

func (s *Framework) pullImage(imageName string) error {
	logger.Instance().Infoln("Realizando el pulling de la imagen", imageName)
	var buf bytes.Buffer
	pullImageOpts := docker.PullImageOptions{Repository: imageName, OutputStream: &buf}
	err := s.client.PullImage(pullImageOpts, s.authConfigs["https://registry.it.lan.com"])
	if err != nil {
		return err
	}

	logger.Instance().Debugln(buf.String())

	if invalidOut := regexp.MustCompile("Pulling .+ Error"); invalidOut.MatchString(buf.String()) {
		return errors.New("Problema al descargar la imagen")
	}

	return nil
}

func bindPort(publish []string) map[docker.Port][]docker.PortBinding {
	portBindings := map[docker.Port][]docker.PortBinding{}

	for _, v := range publish {
		logger.Instance().Debugln("Procesando el bindeo del puerto", v)
		var dp docker.Port
		reflect.ValueOf(&dp).Elem().SetString(v)
		portBindings[dp] = []docker.PortBinding{{}}
	}

	logger.Instance().Debugf("PortBindings %#v", portBindings)

	return portBindings
}

func (s *Framework) DeployService(serviceConfig framework.ServiceConfig, scale int) (*framework.ServiceInformation, error) {
	labels := map[string]string{
		"image_name": serviceConfig.ImageName,
		"image_tag":  serviceConfig.Tag,
	}

	dockerConfig := docker.Config{
		Image:  serviceConfig.ImageName + ":" + serviceConfig.Tag,
		Env:    serviceConfig.Envs,
		Labels: labels,
	}

	sourcetype := "{{.Name}}"
	if serviceConfig.ServiceID != "" {
		sourcetype = serviceConfig.ServiceID
	}

	dockerHostConfig := docker.HostConfig{
		Binds:           []string{"/var/log/service/:/var/log/service/"},
		CPUShares:       int64(serviceConfig.CPUShares),
		PortBindings:    bindPort(serviceConfig.Publish),
		PublishAllPorts: false,
		Privileged:      false,
		LogConfig: docker.LogConfig{
			Type: "syslog",
			Config: map[string]string{
				"tag":             fmt.Sprintf("{{.ImageName}}|%s|{{.ID}}", sourcetype),
				"syslog-facility": "local1",
			},
		},
	}

	if serviceConfig.Memory != 0 {
		dockerHostConfig.Memory = serviceConfig.Memory
	}

	opts := docker.CreateContainerOptions{
		Config:     &dockerConfig,
		HostConfig: &dockerHostConfig}

	err := s.pullImage(opts.Config.Image)
	if err != nil {
		return nil, err
	}

	logger.Instance().Infoln("Creando el contenedor con imagen", opts.Config.Image)
	container, err := s.client.CreateContainer(opts)
	if err != nil {
		return nil, err
	}
	logger.Instance().Infoln("Contenedor creado... Se inicia el proceso de arranque", container.ID)
	err = s.client.StartContainer(container.ID, nil)
	if err != nil {
		switch err.(type) {
		case *docker.NoSuchContainer:
			return nil, err
		case *docker.ContainerAlreadyRunning:
			logger.Instance().Infof("El contenedor %s ya estaba corriendo", container.ID)
			break
		default:
			return nil, err
		}
	}
	logger.Instance().Infoln("Contenedor corriendo... Inspeccionando sus datos", container.Name)
	criteria := &framework.ImageNameAndImageTagRegexpCriteria{
		FullImageNameRegexp: regexp.MustCompile(serviceConfig.ImageName + ":" + serviceConfig.Tag),
	}

	result, err := s.FindServiceInformation(criteria) // TODO FIX
	if err != nil {
		return nil, err
	} else if len(result) == 0 {
		return nil, errors.New("No hay resultados")
	}

	return result[0], nil
}

func (s *Framework) UndeployInstance(containerID string) error {
	remove := false
	var timeout uint = 10
	logger.Instance().Infoln("Se está iniciando el proceso de undeploy del contenedor", containerID)

	// Un valor de 0 sera interpretado como por defecto
	if timeout == 0 {
		timeout = 10
	}

	logger.Instance().Infoln("Deteniendo el contenedor", containerID)
	err := s.client.StopContainer(containerID, timeout)

	if err != nil {
		switch err.(type) {
		case *docker.NoSuchContainer:
			logger.Instance().Infoln("No se encontró el contenedor", containerID)
			return nil
		case *docker.ContainerNotRunning:
			logger.Instance().Infof("El contenedor %s no estaba corriendo", containerID)
			break
		default:
			return err
		}
	}

	if remove {
		logger.Instance().Infoln("Se inició el proceso de remover el contenedor", containerID)
		opts := docker.RemoveContainerOptions{ID: containerID}
		err = s.client.RemoveContainer(opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Framework) DeleteService(serviceId string) error {
	return errors.New("Not implemented yet")
}
