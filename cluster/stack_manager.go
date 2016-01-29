package cluster

import (
	"errors"
	"github.com/latam-airlines/crane/configuration"
	"github.com/latam-airlines/crane/util"
	"github.com/latam-airlines/mesos-framework-factory"
)

type CraneManager interface {
	Deploy(serviceConfig framework.ServiceConfig, instances int, tolerance float64) bool
	FindServiceInformation(string) []*framework.ServiceInformation
	DeployedContainers() []*framework.ServiceInformation
	Rollback()
	DeleteService(string) error
}

type StackManager struct {
	stacks            map[string]StackInterface
	stackNotification chan StackStatus
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
				util.Log.Warnln(err.Error())
				continue
			default:
				return err
			}
		}

		sm.stacks[key] = s
		util.Log.Infof("Cluster %s was configured", key)
	}

	if len(sm.stacks) == 0 {
		return errors.New("Should exist at least one cluster")
	}
	return nil
}

func (sm *StackManager) Deploy(serviceConfig framework.ServiceConfig, instances int, tolerance float64) bool {
	util.Log.Infof("enter deploy stack manager - stacks: %d", len(sm.stacks))

	chanMap := make(map[string]chan StackStatus)

	for stackKey := range sm.stacks {
		ch := make(chan StackStatus)
		chanMap[stackKey] = ch
		go sm.stacks[stackKey].DeployCheckAndNotify(serviceConfig, instances, tolerance, ch)
	}

	//Checking for results on each go routine
	for stackKey, ch := range chanMap {
		if status := <-ch; status == STACK_READY {
			util.Log.Infof("Deploy Process OK on stack %s", stackKey)
		} else {
			util.Log.Errorf("Deploy Process Fails ok stack %s", stackKey)
			sm.Rollback()
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
			util.Log.Errorln(err)
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

func (sm *StackManager) Rollback() {
	util.Log.Infoln("Starting Rollback")
	for stack := range sm.stacks {
		sm.stacks[stack].Rollback()
	}
}

func (sm *StackManager) DeleteService(serviceId string) error {
	util.Log.Infoln("Starting DeleteService")

	chanMap := make(map[string]chan error)

	for stackKey := range sm.stacks {
		ch := sm.stacks[stackKey].DeleteService(serviceId)
		chanMap[stackKey] = ch
	}

	//Checking for results on each go routine
	for stackKey, ch := range chanMap {
		if err := <-ch; err == nil {
			util.Log.Infof("Delete Process OK on stack %s", stackKey)
		} else {
			util.Log.Errorf("Delete Process Fails ok stack %s", stackKey)
			sm.Rollback()
			return err
		}
	}

	return nil
}
