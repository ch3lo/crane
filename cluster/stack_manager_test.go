package cluster

import (
	"reflect"
	"testing"
	"time"

	"github.com/latam-airlines/crane/configuration"
	"github.com/latam-airlines/mesos-framework-factory"
	_ "github.com/latam-airlines/mesos-framework-factory/marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type StackMock struct {
	mock.Mock
	mockId int
}

func (s *StackMock) getServices() []*framework.ServiceInformation {
	s.Called()
	services := make([]*framework.ServiceInformation, 1)
	service := new(framework.ServiceInformation)
	service.ID = "SABRE-SESSION-POOL"
	service.Instances = make([]*framework.Instance, 1)
	service.Instances[0] = new(framework.Instance)
	services[0] = service
	return services
}

func (s *StackMock) undeployInstance(instance string) {
	s.Called()
}

func (s *StackMock) DeployCheckAndNotify(serviceConfig framework.ServiceConfig, instances int, tolerance float64, ch chan int) {
	s.Called(serviceConfig, instances, tolerance, ch)

	if s.mockId == 1 {
		ch <- 0 // ok
	} else {
		ch <- 1 // fails
	}
	return
}

func (s *StackMock) FindServiceInformation(search string) ([]*framework.ServiceInformation, error) {
	s.Called(search)
	services := make([]*framework.ServiceInformation, 1)
	service := new(framework.ServiceInformation)
	service.ID = "SABRE-SESSION-POOL"
	service.Instances = make([]*framework.Instance, 1)
	service.Instances[0] = new(framework.Instance)
	services[0] = service
	return services, nil
}

func (s *StackMock) DeleteService(serviceId string) error {
	s.Called(serviceId)
	return nil
}

func (s *StackMock) Rollback() {
	s.Called()
	return
}

func TestConstructor(t *testing.T) {
	config := &configuration.Configuration{
		Clusters: map[string]configuration.Cluster{
			"local": configuration.Cluster{
				Framework: configuration.Framework{
					"marathon": configuration.Parameters{
						"address":        "http://localhost:8081/v2",
						"deploy-timeout": 30,
					},
				},
			},
		},
	}
	sm, _ := NewStackManager(config)
	assert.NotNil(t, sm, "Instance should be healthy")
	v := reflect.ValueOf(sm).Elem()
	stacks := v.FieldByName("stacks")
	assert.Equal(t, 1, stacks.Len(), "Cli should instantiate at least one stack")
}

func TestDeployMethod(t *testing.T) {
	sm := new(StackManager)
	sm.stacks = make(map[string]StackInterface)
	sm.stackNotification = make(chan StackStatus, 100)

	svc := framework.ServiceConfig{}
	//ch := make(chan int)

	stackMock := new(StackMock)
	stackMock.mockId = 1
	stackMock.On("Rollback").Return().On("DeployCheckAndNotify", svc, 2, 0.0, mock.AnythingOfType("chan int")).WaitUntil(time.After(500 * time.Millisecond)).Return()
	key := ""
	sm.stacks[key] = stackMock
	stackMock = new(StackMock)
	stackMock.mockId = 2
	stackMock.On("Rollback").Return().On("DeployCheckAndNotify", svc, 2, 0.0, mock.AnythingOfType("chan int")).WaitUntil(time.After(500 * time.Millisecond)).Return()
	key = ""
	sm.stacks[key] = stackMock
	sm.Deploy(svc, 2, 0.0)
	stackMock.AssertExpectations(t)

}

func TestDeleteService(t *testing.T) {
	sm := new(StackManager)
	sm.stacks = make(map[string]StackInterface)
	sm.stackNotification = make(chan StackStatus, 100)
	stackMock := new(StackMock)
	key := ""
	serviceId := "serviceId"
	sm.stacks[key] = stackMock
	stackMock.On("DeleteService", serviceId).Return(nil)
	sm.DeleteService(serviceId)
	stackMock.AssertExpectations(t)
}

func TestFindServiceInformation(t *testing.T) {
	sm := new(StackManager)
	sm.stacks = make(map[string]StackInterface)
	sm.stackNotification = make(chan StackStatus, 100)
	stackMock := new(StackMock)
	key := ""
	sm.stacks[key] = stackMock
	search := "search"
	stackMock.On("FindServiceInformation", search).Return(mock.AnythingOfType("[]*framework.ServiceInformation"))
	sm.FindServiceInformation(search)
	stackMock.AssertExpectations(t)
}

func TestRollback(t *testing.T) {
	sm := new(StackManager)
	sm.stacks = make(map[string]StackInterface)
	sm.stackNotification = make(chan StackStatus, 100)
	stackMock := new(StackMock)
	key := ""
	sm.stacks[key] = stackMock
	stackMock.On("Rollback").Return()
	sm.Rollback()
	stackMock.AssertExpectations(t)
}

func TestDeployedContainers(t *testing.T) {
	sm := new(StackManager)
	sm.stacks = make(map[string]StackInterface)
	sm.stackNotification = make(chan StackStatus, 100)
	stackMock := new(StackMock)
	key := ""
	sm.stacks[key] = stackMock
	stackMock.On("getServices").Return(mock.AnythingOfType("[]*framework.ServiceInformation"))
	sm.DeployedContainers()
	stackMock.AssertExpectations(t)
}

func TestInvalidFramework(t *testing.T) {
                config := &configuration.Configuration{
                        Clusters: map[string]configuration.Cluster{
                                "local": {
                                        Framework: configuration.Framework{
                                                "otherFramework": configuration.Parameters{
                                                        "address": "http://localhost:8011",
                                                },
                                        },
                                },
                        },
                }
	_, err := NewStackManager(config)
        assert.NotNil(t, err, "Should return error")
}
