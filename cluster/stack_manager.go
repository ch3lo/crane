package cluster

import (
	"github.com/latam-airlines/crane/util"
	"github.com/latam-airlines/mesos-framework-factory"
	"github.com/latam-airlines/crane/configuration"
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

func NewStackManager(config *configuration.Configuration) CraneManager {
	sm := new(StackManager)
	sm.stacks = make(map[string]StackInterface)
	sm.stackNotification = make(chan StackStatus, 100)
	
	sm.setupStacks(config.Clusters)

	return sm
}

// setupClusters initializes the cluster, mapping the id of the cluster as its key
func (sm *StackManager) setupStacks(config map[string]configuration.Cluster) {
	for key := range config {
		s, err := NewStack(key, sm.stackNotification, config[key])
		if err != nil {
			switch err.(type) {
			case *ClusterDisabled:
				util.Log.Warnln(err.Error())
				continue
			default:
				util.Log.Fatalln(err.Error())
			}
		}

		sm.stacks[key] = s
		util.Log.Infof("Cluster %s was configured", key)
	}

	if len(sm.stacks) == 0 {
		util.Log.Fatalln("Should exist at least one cluster")
	}
}

func (sm *StackManager) Deploy(serviceConfig framework.ServiceConfig, instances int, tolerance float64) bool {
	util.Log.Infof("enter deploy stack manager - stacks: %d", len(sm.stacks))

	chanMap := make(map[string]chan int)

	for stackKey := range sm.stacks {
		ch := make(chan int) // 0 == Ok, 1 == Error
		chanMap[stackKey] = ch
		go sm.stacks[stackKey].DeployCheckAndNotify(serviceConfig, instances, tolerance, ch)
	}

	//Checking for results on each go routine
	for stackKey, ch := range chanMap {
		if <-ch == 0 {
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
	for stack := range sm.stacks {
		sm.stacks[stack].DeleteService(serviceId)
	}
	return nil
}
