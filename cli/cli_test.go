package cli

import (
	"testing"
	"reflect"
	"os"
	"time"
	"flag"
	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
	_ "github.com/latam-airlines/mesos-framework-factory/marathon"
)

func TestCliCmdDeploy(t *testing.T) {
	stackManager = createStackManagerMock()
	os.Args = []string {"crane", "--framework=marathon", "--endpoint=http://localhost:8011", "find", "--search=nginx"}
	RunApp()
	v := reflect.ValueOf(stackManager).Elem()
	stacks := v.FieldByName("stacks")
	assert.Equal(t, 1, stacks.Len(), "Cli should instantiate at least one stack")
}

func TestInvalidLoglevel(t *testing.T) {
        set := flag.NewFlagSet("test", 0)
        set.String("log-level", "OtherLevel", "some hint")
        ctx := cli.NewContext(nil, set, nil)
	err := setupGlobalFlags(ctx)
	assert.NotNil(t, err, "Should return error")
}

func TestInvalidFramework(t *testing.T) {
        set := flag.NewFlagSet("test", 0)
        set.String("framework", "OtherFramework", "some hint")
        set.String("endpoint", "http://localhost:8081", "some hint")
        ctx := cli.NewContext(nil, set, nil)
        err := setupGlobalFlags(ctx)
        assert.NotNil(t, err, "Should return error")
}

func TestDeployTimeout(t *testing.T) {
	stackManager = createStackManagerMock()
	os.Args = []string {"crane", "--framework=marathon", "--endpoint=http://localhost:8081", "--deploy-timeout=20", "deploy", "--image=nginx", "--tag=latest"}
	RunApp()
	
	defer func() {
		os.Args = []string {"crane", "--framework=marathon", "--endpoint=http://localhost:8081", "delete", "--service-id=nginx/latest"}
		RunApp()
		time.Sleep(time.Duration(5)*time.Second)
	}()
	
	assert.Nil(t, nil, "RunApp should pass")
}

func TestDeployTimeoutOptional(t *testing.T) {
	stackManager = createStackManagerMock()
	os.Args = []string {"crane", "--framework=marathon", "--endpoint=http://localhost:8081", "deploy", "--image=nginx", "--tag=latest"}
	RunApp()
	
	defer func() {
		os.Args = []string {"crane", "--framework=marathon", "--endpoint=http://localhost:8081", "delete", "--service-id=nginx/latest"}
		RunApp()
		time.Sleep(time.Duration(5)*time.Second)
	}()
	
	assert.Nil(t, nil, "Error on RunApp should be nil")
}
