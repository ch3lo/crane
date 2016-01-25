package cluster

import (
	"github.com/latam-airlines/crane/util"
	"github.com/latam-airlines/mesos-framework-factory"
)

type CraneManager interface {
	AppendStack(fh framework.Framework)
	Deploy(serviceConfig framework.ServiceConfig, instances int, tolerance float64) bool
	FindServiceInformation(string) []*framework.ServiceInformation
	DeployedContainers () []*framework.ServiceInformation
	Rollback()
	DeleteService(string) error
}

type StackManager struct {
	stacks            map[string]StackInterface
	stackNotification chan StackStatus
}

func NewStackManager() CraneManager {
	sm := new(StackManager)
	sm.stacks = make(map[string]StackInterface)
	sm.stackNotification = make(chan StackStatus, 100)

	return sm
}

func (sm *StackManager) createId() string {
	i := 0
	for {
		key := util.Letter(i)
		exist := false

		for k := range sm.stacks {
			if k == key {
				exist = true
			}
		}

		if !exist {
			return key
		}
		i++
	}
}

func (sm *StackManager) AppendStack(fh framework.Framework) {
	key := sm.createId()
	util.Log.Infof("API configurada y mapeada a la llave %s", key)
	sm.stacks[key] = NewStack(key, sm.stackNotification, fh)
}

func (sm *StackManager) Deploy(serviceConfig framework.ServiceConfig, instances int, tolerance float64) bool {
	util.Log.Infof("enter deploy stack manager - stacks: %d", len(sm.stacks))
	
	chanMap := make(map[string]chan int)

	for stackKey, _ := range sm.stacks {
		ch := make(chan int) // 0 == Ok, 1 == Error
		chanMap[stackKey] = ch
		go sm.stacks[stackKey].DeployCheckAndNotify(serviceConfig, instances, tolerance, ch)
	}
	
	//Checking for results on each go routine
	for stackKey, ch  := range chanMap {
		if (<-ch == 0) {
			util.Log.Infof("Proceso de deploy OK en stack %s", stackKey)
		} else {
			util.Log.Errorf("Proceso de deploy Falló en stack %s", stackKey)
		}
	}
	return true
}

func (sm *StackManager) FindServiceInformation(search string) []*framework.ServiceInformation {
	allServices := make([]*framework.ServiceInformation, 0)
        for stack, _ := range sm.stacks {
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

func (sm *StackManager) DeployedContainers () []*framework.ServiceInformation {
        allServices := make([]*framework.ServiceInformation, 0)
        for stack, _ := range sm.stacks {
                services := sm.stacks[stack].getServices()
                if services != nil || len(services) != 0 {
                        allServices = append(allServices, services...)
                }
        }
        return allServices
	
}

func (sm *StackManager) Rollback() {
       util.Log.Infoln("Starting Rollback")
       for stack, _ := range sm.stacks {
               sm.stacks[stack].Rollback()
       }
}

func (sm *StackManager) DeleteService(serviceId string) error {
	util.Log.Infoln("Starting DeleteService")
	for stack, _ := range sm.stacks {
		sm.stacks[stack].DeleteService(serviceId)
	}
	return nil
}
