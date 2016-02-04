package marathon

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gambol99/go-marathon"
	"github.com/latam-airlines/mesos-framework-factory"
	"github.com/latam-airlines/mesos-framework-factory/factory"
	"github.com/latam-airlines/mesos-framework-factory/logger"
	"github.com/latam-airlines/mesos-framework-factory/utils"
	"strings"
	"time"
)

const frameworkID = "marathon"

func init() {
	factory.Register(frameworkID, &marathonCreator{})
}

// marathonCreator implements factory.FrameworkFactory
type marathonCreator struct{}

func (factory *marathonCreator) Create(params map[string]interface{}) (framework.Framework, error) {
	address, ok := params["address"]
	if !ok || fmt.Sprint(address) == "" {
		return nil, errors.New("Parameter address does not exist")
	}

	deployTimeout, ok := params["deploy-timeout"].(int)

	if !ok {
		return nil, errors.New("Parameter deploy-timeout does not exist")
	}
	authUser := utils.ExtractString(params, "basic-auth-user")
	authPwd := utils.ExtractString(params, "basic-auth-pwd")
	if authUser != "" && authPwd == "" {
		return nil, errors.New("Parameter basic-auth-pwd does not exist")
	}

	parameters := &Parameters{
		EndpointUrl:           fmt.Sprint(address),
		DeployTimeout:         deployTimeout,
		HTTPBasicAuthUser:     authUser,
		HTTPBasicAuthPassword: authPwd,
		DockerCfg:             utils.ExtractString(params, "docker-cfg"),
	}

	helper, err := NewMarathon(parameters)

	if err == nil {
		helper.healthCheckConf = &framework.HealthCheck{
			GracePeriod:            utils.ExtractNaturalNumber(params, "health-check-grace-period"),
			Interval:               utils.ExtractNaturalNumber(params, "health-check-interval"),
			Timeout:                utils.ExtractNaturalNumber(params, "health-check-timeout"),
			MaxConsecutiveFailures: utils.ExtractNaturalNumber(params, "health-check-max-consecutive-failures"),
		}
	}

	return helper, err
}

type Marathon struct {
	client          marathon.Marathon
	endpointUrl     string
	deployTimeout   int
	dockerCfg       string
	healthCheckConf *framework.HealthCheck
}

type Parameters struct {
	EndpointUrl           string
	DeployTimeout         int
	HTTPBasicAuthUser     string
	HTTPBasicAuthPassword string
	DockerCfg             string
}

func NewMarathon(params *Parameters) (*Marathon, error) {
	helper := new(Marathon)
	endpointUrl := utils.ValidateEndpoint(params.EndpointUrl)

	helper.endpointUrl = endpointUrl
	helper.deployTimeout = params.DeployTimeout
	config := marathon.NewDefaultConfig()
	config.URL = endpointUrl
	if params.HTTPBasicAuthUser != "" {
		config.HTTPBasicAuthUser = params.HTTPBasicAuthUser
		config.HTTPBasicPassword = params.HTTPBasicAuthPassword
	}
	client, err := marathon.NewClient(config)

	if err != nil {
		return nil, err
	}

	helper.client = client
	helper.dockerCfg = params.DockerCfg
	return helper, nil
}

func (helper *Marathon) SetClient(client marathon.Marathon) {
	helper.client = client
}

func (helper *Marathon) GetEndpointUrl() string {
	return helper.endpointUrl
}

func (helper *Marathon) FindServiceInformation(criteria framework.ServiceInformationCriteria) ([]*framework.ServiceInformation, error) {
	//by default this method does not return task-info
	apps, err := helper.client.Applications(nil)
	if err != nil {
		return nil, err
	} else {
		if apps == nil {
			return make([]*framework.ServiceInformation, 0), nil
		} else {
			allServices := make([]*framework.ServiceInformation, len(apps.Apps))
			for i, app := range apps.Apps {
				allServices[i] = helper.getServiceInformationFromApp(&app)
			}
			filteredServices := criteria.MeetCriteria(allServices)
			if filteredServices == nil || len(filteredServices) == 0 {
				return nil, errors.New("No services found")
			}
			// request task info of app, that is why we do the loop again
			services := make([]*framework.ServiceInformation, len(filteredServices))
			for i, service := range filteredServices {
				app, err := helper.client.Application(service.ID)
				if err != nil {
					return nil, errors.New("Error listing filtered services")
				}
				services[i] = helper.getServiceInformationFromApp(app)
			}
			return services, nil
		}
	}
}

func (m *Marathon) getServiceInformationFromApp(app *marathon.Application) *framework.ServiceInformation {
	service := framework.ServiceInformation{ImageTag: "latest"}
	service.ID = app.ID
	imageInfo := strings.Split(app.Container.Docker.Image, ":")
	service.ImageName = imageInfo[0]
	if len(imageInfo) > 1 {
		service.ImageTag = imageInfo[1]
	}
	service.Instances = m.getInstancesFromTasks(app.Tasks, app.Container.Docker.PortMappings)
	return &service
}

/* Sets global health check config to the service deploy config */
func copyGlobalHealthCheckToServiceCfg(healthCheck *framework.HealthCheck, config *framework.ServiceConfig) {
	if config.HealthCheckConfig != nil {
		config.HealthCheckConfig.GracePeriod = healthCheck.GracePeriod
		config.HealthCheckConfig.Interval = healthCheck.Interval
		config.HealthCheckConfig.MaxConsecutiveFailures = healthCheck.MaxConsecutiveFailures
		config.HealthCheckConfig.Timeout = healthCheck.Timeout
	}
}

func (helper *Marathon) createService(config *framework.ServiceConfig, instances int) (*framework.ServiceInformation, error) {
	config.DockerCfg = helper.dockerCfg
	copyGlobalHealthCheckToServiceCfg(helper.healthCheckConf, config)
	app := translateServiceConfig(config, instances)
	appResult, err := helper.client.CreateApplication(app)
	if err != nil {
		return nil, err
	} else {

		logger.Instance().Debugln("#### appResult marshall ####")
		jsonresult, _ := json.Marshal(appResult)
		logger.Instance().Debugf("App jsonresult:  \n\n %s", string(jsonresult))

		deployErr := helper.client.WaitOnApplication(app.ID, time.Duration(helper.deployTimeout)*time.Second)

		if deployErr != nil {
			logger.Instance().Errorf("Failed to Create the application: %s, error: %s \n", app.ID, deployErr)
			logger.Instance().Infof("Executing Rollback: Deleting App with ID %s", app.ID)
			helper.DeleteService(app.ID)
			return nil, deployErr
		}

		app, err := helper.client.Application(app.ID)
		if err != nil {
			return nil, err
		} else {
			return helper.getServiceInformationFromApp(app), nil
		}
	}
}
func (helper *Marathon) DeployService(config framework.ServiceConfig, instances int) (*framework.ServiceInformation, error) {
	apps, err := helper.client.ListApplications(nil)
	if err != nil {
		return nil, err
	}
	if !helper.containsApp(apps, config.ServiceID) {
		return helper.createService(&config, instances)
	} else {
		return helper.scaleService(config.ServiceID, instances)
	}
}

func (helper *Marathon) scaleService(id string, instances int) (*framework.ServiceInformation, error) {
	deploymentId, err := helper.client.ScaleApplicationInstances(id, instances, true)
	if err != nil {
		logger.Instance().Errorf("Failed to Scale the application: %s, error: %s", id, err)
		return nil, err
	} else {
		app, err := helper.client.Application(id)
		if err != nil {
			return nil, err
		} else {
			serviceInformation := helper.getServiceInformationFromApp(app)
			serviceInformation.Instances = helper.getInstancesByVersion(app.Tasks, app.Container.Docker.PortMappings, deploymentId.Version)
			return serviceInformation, nil
		}
	}
}

func (helper *Marathon) DeleteService(id string) error {
	_, err := helper.client.DeleteApplication(id)
	if err != nil {
		logger.Instance().Errorf("Failed to Delete the application: %s, error: %s", id, err)
	}
	return err
}

func (scheduler *Marathon) UndeployInstance(id string) error {
	return errors.New("Not implemented yet")
}

func (helper *Marathon) getInstancesFromTasks(tasks []*marathon.Task, dockerPortMappings []*marathon.PortMapping) []*framework.Instance {
	instances := make([]*framework.Instance, len(tasks))
	if tasks == nil || len(tasks) == 0 {
		return nil
	}
	for i, task := range tasks {
		instance := framework.Instance{}
		instance.ID = task.ID
		instance.Host = task.Host
		//instance.ContainerName
		instance.Ports = helper.buildInstancePorts(dockerPortMappings, task.Ports)
		instances[i] = &instance
	}
	return instances
}

func (helper *Marathon) buildInstancePorts(dockerPortMappings []*marathon.PortMapping, taskPorts []int) map[string]framework.InstancePort {
	if dockerPortMappings == nil || len(dockerPortMappings) == 0 || taskPorts == nil || len(taskPorts) == 0 {
		return nil
	}

	ports := make(map[string]framework.InstancePort, len(taskPorts))

	for i, port := range taskPorts {
		instancePort := framework.InstancePort{}
		instancePort.Type = framework.NewInstancePortType(dockerPortMappings[i].Protocol)
		instancePort.Internal = int64(port)
		instancePort.Publics = []int64{int64(dockerPortMappings[i].ContainerPort)}
		//instancePort.Advertise =
		ports[string(instancePort.Publics[0])+"/"+string(instancePort.Type)] = instancePort
	}
	return ports
}

func (helper *Marathon) getInstancesByVersion(tasks []*marathon.Task, dockerPortMappings []*marathon.PortMapping, version string) []*framework.Instance {
	for i, task := range tasks {
		if task.Version != version {
			tasks = append(tasks[:i], tasks[i+1:]...)
		}
	}
	return helper.getInstancesFromTasks(tasks, dockerPortMappings)
}

func (m *Marathon) containsApp(apps []string, search string) bool {
	for _, a := range apps {
		if a == search {
			return true
		}
	}
	return false
}
func (s *Marathon) ID() string {
	return frameworkID
}
