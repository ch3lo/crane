package cluster

import (
	"errors"

	"github.com/latam-airlines/crane/configuration"
	"github.com/latam-airlines/crane/logger"
	"github.com/latam-airlines/mesos-framework-factory"
)

type CraneManager interface {
	Deploy(serviceConfig framework.ServiceConfig, instances int, tolerance float64) bool
	FindServiceInformation(string) []*framework.ServiceInformation
	DeployedContainers() []*framework.ServiceInformation
	Rollback(string, string)
	DeleteService(string) error
}

type StackManager struct {
	stacks            map[string]StackInterface
	stackNotification chan StackStatus
}

type ServiceInfoStatus struct {
	serviceInfo *framework.ServiceInformation
	status      StackStatus
}

func NewStackManager(config *configuration.Configuration) (CraneManager, error) {
	sm := new(StackManager)
	sm.stacks = make(map[string]StackInterface)
	sm.stackNotification = make(chan StackStatus, 100)

	err := sm.setupStacks(config.Clusters)
	if err != nil {
		return nil, err
	}

	return sm, nil
}

// setupClusters initializes the cluster, mapping the id of the cluster as its key
func (sm *StackManager) setupStacks(config map[string]configuration.Cluster) error {
	for key := range config {
		s, err := NewStack(key, sm.stackNotification, config[key])
		if err != nil {
			switch err.(type) {
			case *ClusterDisabled:
				logger.Instance().Warnln(err.Error())
				continue
			default:
				return err
			}
		}

		sm.stacks[key] = s
		logger.Instance().Infof("Cluster %s was configured", key)
	}

	if len(sm.stacks) == 0 {
		return errors.New("Should exist at least one cluster")
	}
	return nil
}

func (sm *StackManager) Deploy(serviceConfig framework.ServiceConfig, instances int, tolerance float64) bool {
	logger.Instance().Infof("enter deploy stack manager - stacks: %d", len(sm.stacks))

	chanMap := make(map[string]chan *ServiceInfoStatus)

	for stackKey := range sm.stacks {
		ch := make(chan *ServiceInfoStatus)
		chanMap[stackKey] = ch
		go sm.stacks[stackKey].DeployCheckAndNotify(serviceConfig, instances, tolerance, ch)
	}

	//Checking for results on each go routine
	for stackKey, ch := range chanMap {
		if serviceInfoStatus := <-ch; serviceInfoStatus.status == STACK_READY {
			logger.Instance().Infof("Deploy Process OK on stack %s, status %d", stackKey, serviceInfoStatus.status)
		} else {
			logger.Instance().Errorf("Deploy Process Fails on stack %s", stackKey)

			if serviceInfoStatus.serviceInfo != nil {
				sm.Rollback(serviceInfoStatus.serviceInfo.ID, serviceInfoStatus.serviceInfo.Version)
			}
			return false
		}
	}
	return true
}

func (sm *StackManager) FindServiceInformation(search string) []*framework.ServiceInformation {
	allServices := make([]*framework.ServiceInformation, 0)
	for stack := range sm.stacks {
		services, err := sm.stacks[stack].FindServiceInformation(search)
		if err != nil {
			logger.Instance().Errorln(err)
		}
		if services != nil || len(services) != 0 {
			allServices = append(allServices, services...)
		}
	}
	return allServices
}

func (sm *StackManager) DeployedContainers() []*framework.ServiceInformation {
	var allServices []*framework.ServiceInformation
	for stack := range sm.stacks {
		services := sm.stacks[stack].getServices()
		if services != nil || len(services) != 0 {
			allServices = append(allServices, services...)
		}
	}
	return allServices
}

func (sm *StackManager) Rollback(appId, previousVersion string) {
	logger.Instance().Infoln("Starting Rollback")
	for stack := range sm.stacks {
		sm.stacks[stack].Rollback(appId, previousVersion)
	}
}

func (sm *StackManager) DeleteService(serviceId string) error {
	logger.Instance().Infoln("Starting DeleteService")

	chanMap := make(map[string]chan error)

	for stackKey := range sm.stacks {
		ch := sm.stacks[stackKey].DeleteService(serviceId)
		chanMap[stackKey] = ch
	}

	//Checking for results on each go routine
	for stackKey, ch := range chanMap {
		if err := <-ch; err == nil {
			logger.Instance().Infof("Delete Process OK on stack %s", stackKey)
		} else {
			logger.Instance().Errorf("Delete Process Fails ok stack %s", stackKey)
			// XXX: Se elimina Rollback(), se debe implementar Retry Configurable PAAS-593
			return err
		}
	}

	return nil
}
