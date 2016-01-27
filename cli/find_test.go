package cli

import (
        "testing"
	"flag"
        "github.com/stretchr/testify/assert"
	"github.com/codegangsta/cli"
	"github.com/latam-airlines/mesos-framework-factory"
	"github.com/latam-airlines/crane/cluster"
)

type StackManagerMock struct {}
func (sm *StackManagerMock) buildServiceDummyList() []*framework.ServiceInformation {
        services := make([]*framework.ServiceInformation, 1)
        service := new(framework.ServiceInformation)
        service.ID = "SABRE-SESSION-POOL"
        service.Instances = make([]*framework.Instance, 1)
        service.Instances[0] = new(framework.Instance)
	service.Instances[0].ID = "instance id"
        services[0] = service
	return services
}
func (sm *StackManagerMock) FindServiceInformation(search string) []*framework.ServiceInformation {
	return sm.buildServiceDummyList()
}
func (sm *StackManagerMock) AppendStack(fh framework.Framework) {}
func (sm *StackManagerMock) Deploy(serviceConfig framework.ServiceConfig, instances int, tolerance float64) bool { return true }
func (sm *StackManagerMock) DeployedContainers () []*framework.ServiceInformation { return sm.buildServiceDummyList() }
func (sm *StackManagerMock) Rollback() {}
func (sm *StackManagerMock) DeleteService(string) error {return nil}

func createStackManagerMock() cluster.CraneManager {
	return new(StackManagerMock)
}

func TestFindFlags(t *testing.T) {
	flags := findFlags()
	stringFlag, _ := flags[0].(cli.StringFlag)
	assert.Equal(t, "search", stringFlag.Name, "Should be search")
}

func TestFindBefore(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("search", "SABRE-SESSION-POOL-v1", "some hint")
	ctx := cli.NewContext(nil, set, nil)
        err := findBefore(ctx)
        assert.Nil(t, err, "Should be nil")
}

func TestFindBeforeError(t *testing.T) {
        set := flag.NewFlagSet("test", 0)
        set.String("searchX", "SABRE-SESSION-POOL-v1", "some hint")
        ctx := cli.NewContext(nil, set, nil)
        err := findBefore(ctx)
        assert.NotNil(t, err, "Should throw error")
}

func TestFindCmd(t *testing.T) {
	stackManager = createStackManagerMock()
        set := flag.NewFlagSet("test", 0)
        set.String("search", "SABRE-SESSION-POOL-v1", "some hint")
        ctx := cli.NewContext(nil, set, nil)
        findCmd(ctx)
}
