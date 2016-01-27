package cluster

import (
	"fmt"
	"github.com/Pallinder/go-randomdata"
	log "github.com/Sirupsen/logrus"
	"github.com/latam-airlines/crane/util"
	"github.com/latam-airlines/mesos-framework-factory"
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
	createId() string
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

func NewStack(stackKey string, stackNofitication chan<- StackStatus, fh framework.Framework) StackInterface {
	s := new(Stack)
	s.id = stackKey
	s.stackNofitication = stackNofitication
	s.frameworkApiHelper = fh
	s.serviceIdNotification = make(chan string, 1000)

	s.log = util.Log.WithFields(log.Fields{
		"stack": stackKey,
	})

	return s
}

func (s *Stack) getServices() []*framework.ServiceInformation {
	return s.services
}

func (s *Stack) createId() string {
	for {
		key := s.id + "_" + randomdata.Adjective()
		exist := false

		for _, srv := range s.services {
			if srv.ID == key {
				exist = true
			}
		}

		if !exist {
			return key
		}
	}
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
