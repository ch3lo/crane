package cluster

import (
	"github.com/Pallinder/go-randomdata"
	log "github.com/Sirupsen/logrus"
	"github.com/latam-airlines/crane/util"
	"fmt"
	"regexp"
	"github.com/latam-airlines/mesos-framework-factory"
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

type Stack struct {
	id                    string
	frameworkApiHelper    framework.Framework
	services              []*framework.ServiceInformation
	serviceIdNotification chan string
	stackNofitication     chan<- StackStatus
	log                   *log.Entry
}

func NewStack(stackKey string, stackNofitication chan<- StackStatus, fh framework.Framework) *Stack {
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


func (s *Stack) DeployCheckAndNotify(serviceConfig framework.ServiceConfig, instances int, tolerance float64) {
	_, err := s.frameworkApiHelper.DeployService(serviceConfig, instances)
	if err != nil {
		fmt.Println(err)
	}
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
	services , err := s.frameworkApiHelper.FindServiceInformation(&framework.ImageNameAndImageTagRegexpCriteria{regexp.MustCompile(search)})
	if err != nil {
		return nil , err
	}
	s.services = services
	return s.services, nil
}

func (s *Stack) DeleteService(serviceId string) error {
	return s.frameworkApiHelper.DeleteService(serviceId)
}
