package cluster

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
	"math/rand"
	"reflect"
	"github.com/latam-airlines/mesos-framework-factory"
	"github.com/latam-airlines/mesos-framework-factory/factory"
)

type StackMock struct {
	mock.Mock
}

func (s *StackMock) getServices() []*framework.ServiceInformation { 
	s.Called()
	return nil
}

func (s *StackMock) createId() string { 
	s.Called()
	return ""
}

func (s *StackMock) undeployInstance(instance string) {
	s.Called()
}

func (s *StackMock) DeployCheckAndNotify(serviceConfig framework.ServiceConfig, instances int, tolerance float64, ch chan int) {
	s.Called(serviceConfig, instances, tolerance, ch)
	time.Sleep(100 * time.Millisecond)
	
	s1 := rand.NewSource(time.Now().UnixNano())
    rnd := rand.New(s1)
	ch <- rnd.Intn(2) // fails randomly
	return
}

func (s *StackMock) FindServiceInformation(search string) ([]*framework.ServiceInformation, error) {
	s.Called()
	return nil, nil
}

func (s *StackMock) DeleteService(serviceId string) error {
	s.Called()
	return nil
}

func (s *StackMock) Rollback() {
	s.Called()
}

func TestConstructor(t *testing.T) {
	sm := NewStackManager()
	assert.True(t, sm != nil, "Instance should be healthy")
	params := make(map[string]interface{})
	params["address"] = "http://localhost:8081/v2"
	helper, _ := factory.Create("marathon", params)
	sm.AppendStack(helper)
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
	stackMock.On("DeployCheckAndNotify", svc, 2, 0.0, mock.AnythingOfType("chan int")).WaitUntil(time.After(500*time.Millisecond)).Return()
	key := sm.createId()
	sm.stacks[key] = stackMock
	stackMock = new(StackMock)
	stackMock.On("DeployCheckAndNotify", svc, 2, 0.0, mock.AnythingOfType("chan int")).WaitUntil(time.After(500*time.Millisecond)).Return()
	key = sm.createId()
	sm.stacks[key] = stackMock
	sm.Deploy(svc, 2, 0.0)
	
}
