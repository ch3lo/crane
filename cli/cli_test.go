package cli

import (
	"flag"
	"github.com/codegangsta/cli"
	_ "github.com/latam-airlines/mesos-framework-factory/marathon"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestCliCmdDeploy(t *testing.T) {
	ctx := cli.NewContext(nil, createStandardTestFlagSet(), nil)
	err := setupGlobalFlags(ctx)
	assert.Nil(t, err, "Search should work")
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
	endpointFlag := createEndpointSliceFlag()
	endpointFlag.Apply(set)
	set.String("framework", "OtherFramework", "some hint")
	set.String("log-level", "info", "some hint")
	set.String("log-output", "console", "")
	set.String("log-formatter", "text", "")
	ctx := cli.NewContext(nil, set, nil)
	err := setupGlobalFlags(ctx)
	assert.NotNil(t, err, "Should return error")
}

func TestDeployTimeout(t *testing.T) {
	set := createStandardTestFlagSet()
	set.Int("deploy-timeout", 20, "")
	ctx := cli.NewContext(nil, set, nil)
	err := setupGlobalFlags(ctx)
	assert.Nil(t, err, "Search should work")
}

func createEndpointSliceFlag() cli.StringSliceFlag {
	epSlice := new(cli.StringSlice)
	epSlice.Set("http://localhost:8081")

	endpointFlag := cli.StringSliceFlag{
		Name:  "endpoint, ep",
		Value: epSlice,
	}
	return endpointFlag
}

func createStandardTestFlagSet() *flag.FlagSet {
	set := flag.NewFlagSet("test", 0)
	set.String("framework", "marathon", "")
	set.String("log-level", "info", "")
	set.String("log-output", "console", "")
	set.String("log-formatter", "text", "")

	endpointFlag := createEndpointSliceFlag()
	endpointFlag.Apply(set)
	return set
}
