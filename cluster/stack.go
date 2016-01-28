package cluster

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/latam-airlines/crane/configuration"
	"github.com/latam-airlines/crane/util"
	"github.com/latam-airlines/mesos-framework-factory"
	"github.com/latam-airlines/mesos-framework-factory/factory"
	"regexp"
)

type StackStatus int

const (
	STACK_READY StackStatus = 1 + iota
	STACK_FAILED
)

var stackStatus = [...]string{
	"STACK_READY",
	"STACK_FAILED",
}

func (s StackStatus) String() string {
	return stackStatus[s-1]
}

type StackInterface interface {
	getServices() []*framework.ServiceInformation
	undeployInstance(instance string)
	DeployCheckAndNotify(serviceConfig framework.ServiceConfig, instances int, tolerance float64, ch chan int)
	FindServiceInformation(search string) ([]*framework.ServiceInformation, error)
	DeleteService(serviceId string) error
	Rollback()
}

type Stack struct {
	id                    string
	frameworkApiHelper    framework.Framework
	services              []*framework.ServiceInformation
	serviceIdNotification chan string
	stackNofitication     chan<- StackStatus
	log                   *log.Entry
}

func NewStack(stackKey string, stackNofitication chan<- StackStatus, config configuration.Cluster) (StackInterface, error) {
	if config.Disabled {
		return nil, &ClusterDisabled{Name: stackKey}
	}

	clusterScheduler, err := factory.Create(config.Framework.Type(), config.Framework.Parameters())
	if err != nil {
		return nil, fmt.Errorf("Error creating framework %s in %s. %s", config.Framework.Type(), stackKey, err.Error())
	}
	s := new(Stack)
	s.id = stackKey
	s.stackNofitication = stackNofitication
	s.frameworkApiHelper = clusterScheduler
	s.serviceIdNotification = make(chan string, 1000)

	util.Log.WithFields(log.Fields{
		"stack": stackKey,
	}).Infof("A new framework was created: %s", config.Framework.Type())

	return s, nil
}

func (s *Stack) getServices() []*framework.ServiceInformation {
	return s.services
}

func (s *Stack) DeployCheckAndNotify(serviceConfig framework.ServiceConfig, instances int, tolerance float64, ch chan int) {
	service, err := s.frameworkApiHelper.DeployService(serviceConfig, instances)
	if err != nil {
		ch <- 1 // error
		fmt.Println(err)
	} else {
		ch <- 0 // success
	}
	services := make([]*framework.ServiceInformation, 0)
	services = append(services, service)
	s.services = services
}

func (s *Stack) setStatus(status StackStatus) {
	s.stackNofitication <- status
}

func (s *Stack) undeployInstance(instance string) {
	s.frameworkApiHelper.UndeployInstance(instance)
}

func (s *Stack) Rollback() {
	s.log.Infof("Comenzando Rollback en el Stack")
}

func (s *Stack) FindServiceInformation(search string) ([]*framework.ServiceInformation, error) {
	services, err := s.frameworkApiHelper.FindServiceInformation(&framework.ImageNameAndImageTagRegexpCriteria{regexp.MustCompile(search)})
	if err != nil {
		return nil, err
	}
	s.services = services
	return s.services, nil
}

func (s *Stack) DeleteService(serviceId string) error {
	return s.frameworkApiHelper.DeleteService(serviceId)
}
