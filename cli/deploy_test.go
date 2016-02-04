package cli

import (
	"flag"
	"github.com/codegangsta/cli"
	"github.com/latam-airlines/mesos-framework-factory"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createFlagSetWithMandatoryFlags() *flag.FlagSet {
	set := flag.NewFlagSet("test", 0)
	set.String("service-id", "MyServiceId", "")
	set.String("image", "someImage", "")
	set.String("tag", "someTag", "")
	return set
}

func TestApplyPortsNil(t *testing.T) {
	cfg := new(framework.ServiceConfig)
	applyPorts(nil, cfg)
	assert.Equal(t, "8080/tcp", cfg.Publish[0], "Should add \"8080/tcp\"")
}

func TestApplyPortsWithDefault(t *testing.T) {
	cfg := new(framework.ServiceConfig)
	applyPorts([]string{"8080/tcp"}, cfg)
	assert.Equal(t, "8080/tcp", cfg.Publish[0], "Should add \"8080/tcp\"")
}

func TestApplyPortsWithExtras(t *testing.T) {
	cfg := new(framework.ServiceConfig)
	applyPorts([]string{"8080/tcp", "443/tcp"}, cfg)
	assert.Equal(t, "8080/tcp", cfg.Publish[0], "Should add \"8080/tcp\"")
	assert.Equal(t, "443/tcp", cfg.Publish[1], "Should add \"443/tcp\"")
}

func TestApplyPortsWrongFormat(t *testing.T) {
	cfg := new(framework.ServiceConfig)
	err := applyPorts([]string{"8080+tcp"}, cfg)
	assert.NotNil(t, err, "Should throw error")

	cfg = new(framework.ServiceConfig)
	err = applyPorts([]string{"xyz/tcp"}, cfg)
	assert.NotNil(t, err, "Should throw error")

	cfg = new(framework.ServiceConfig)
	err = applyPorts([]string{"udp/8080"}, cfg)
	assert.NotNil(t, err, "Should throw error")
}

func TestDeployBeforeError(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.NotNil(t, err, "Should throw error service-id empty")

	set.String("service-id", "MyServiceId", "")
	ctx = cli.NewContext(nil, set, nil)
	err = deployBefore(ctx)
	assert.NotNil(t, err, "Should throw error image empty")

	set.String("image", "someImage", "")
	ctx = cli.NewContext(nil, set, nil)
	err = deployBefore(ctx)
	assert.NotNil(t, err, "Should throw error tag empty")

	set.String("tag", "someTag", "")
	set.String("memory", "some memory", "")
	ctx = cli.NewContext(nil, set, nil)
	err = deployBefore(ctx)
	assert.NotNil(t, err, "Should throw error memory empty")

	set = createFlagSetWithMandatoryFlags()
	set.String("memory", "512", "")
	envSlice := new(cli.StringSlice)
	envSlice.Set("/tmp/bla.txt")

	envFlag := cli.StringSliceFlag{
		Name:  "env-file",
		Value: envSlice,
	}
	envFlag.Apply(set)
	ctx = cli.NewContext(nil, set, nil)
	err = deployBefore(ctx)
	assert.NotNil(t, err, "Should throw error file does not exist")
}

func TestDeployBefore(t *testing.T) {
	set := createFlagSetWithMandatoryFlags()
	set.String("memory", "512", "")
	envSlice := new(cli.StringSlice)
	envSlice.Set("deploy.go")

	envFlag := cli.StringSliceFlag{
		Name:  "env-file",
		Value: envSlice,
	}
	envFlag.Apply(set)
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.Nil(t, err, "Should pass without any error")
}

func TestDeployCmd(t *testing.T) {
	stackManager = createStackManagerMock()
	set := createFlagSetWithMandatoryFlags()
	set.String("memory", "512", "")
	envFileSlice := new(cli.StringSlice)
	envFileSlice.Set("deploy.go")

	envFileFlag := cli.StringSliceFlag{
		Name:  "env-file",
		Value: envFileSlice,
	}
	envFileFlag.Apply(set)
	envSlice := new(cli.StringSlice)
	envSlice.Set("SOME_ENV_VAR")

	envFlag := cli.StringSliceFlag{
		Name:  "env",
		Value: envSlice,
	}
	envFlag.Apply(set)

	endpointFlag := createEndpointSliceFlag()
	endpointFlag.Apply(set)
	ctx := cli.NewContext(nil, set, nil)

	deployCmd(ctx)
}

func TestCpuFlag(t *testing.T) {
	set := createFlagSetWithMandatoryFlags()
	set.Float64("cpu", 0.25, "usage")
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.Nil(t, err, "Should be fine")
}

func TestCpuFlagNegative(t *testing.T) {
	set := createFlagSetWithMandatoryFlags()
	set.Float64("cpu", -2.1, "usage")
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.NotNil(t, err, "Should fail")
}

func TestCpuFlagMarathonWrongRange(t *testing.T) {
	set := createFlagSetWithMandatoryFlags()
	set.String("framework", "marathon", "usage") // Fix this: framework flag does not exist anymore
	set.Float64("cpu", 1.1, "usage")
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.NotNil(t, err, "Should fail")
}

func TestCpuFlagSwarmWrongRange(t *testing.T) {
	set := createFlagSetWithMandatoryFlags()
	set.String("framework", "swarm", "usage") // Fix this: framework flag does not exist anymore
	set.Float64("cpu", 1025, "usage")
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.NotNil(t, err, "Should fail")
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

func TestMinimumHealthCapacityFlag(t *testing.T) {
	set := createFlagSetWithMandatoryFlags()
	set.Float64("minimumHealthCapacity", 1.0, "usage")
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.Nil(t, err, "Should be fine")
}

func TestMaximumOverCapacityFlag(t *testing.T) {
	set := createFlagSetWithMandatoryFlags()
	set.Float64("maximumOverCapacity", 0.2, "usage")
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.Nil(t, err, "Should be fine")
}

func TestMinimumHealthCapacityFlagOutRange(t *testing.T) {
	set := createFlagSetWithMandatoryFlags()
	set.Float64("minimumHealthCapacity", 1.1, "usage")
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.NotNil(t, err, "Should fail")
}

func TestMaximumOverCapacityFlagOutRange(t *testing.T) {
	set := createFlagSetWithMandatoryFlags()
	set.Float64("maximumOverCapacity", -0.2, "usage")
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.NotNil(t, err, "Should fail")
}

func TestHealthCheckFlag(t *testing.T) {
	set := createFlagSetWithMandatoryFlags()
	set.String("health-check-path", "/v0/healhty", "usage")
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.Nil(t, err, "Should be fine")
}

func TestApplyKeyValSliceFlagNil(t *testing.T) {
	err := applyKeyValSliceFlag(nil, nil)
	assert.Nil(t, err, "Should do nothing")
}

func TestApplyKeyValSliceFlagConstraints(t *testing.T) {
	constraints := make([]string, 2)
	constraints[0] = "key1=val1"
	constraints[1] = "key2=val2"
	cfg := new(framework.ServiceConfig)
	applyKeyValSliceFlag(constraints, func(configMap map[string]string) {
		if configMap != nil && len(configMap) != 0 {
			cfg.Constraints = configMap
		}
	})
	assert.Equal(t, "val1", cfg.Constraints["key1"], "Should contain key1")
	assert.Equal(t, "val2", cfg.Constraints["key2"], "Should contain key2")
}

func TestApplyKeyValSliceFlagError(t *testing.T) {
	constraints := make([]string, 1)
	constraints[0] = "key2;val2"
	cfg := new(framework.ServiceConfig)
	err := applyKeyValSliceFlag(constraints, func(configMap map[string]string) {
		if configMap != nil && len(configMap) != 0 {
			cfg.Constraints = configMap
		}
	})
	assert.NotNil(t, err, "Should fail")
}

func TestApplyKeyValSliceFlagLabels(t *testing.T) {
	labels := make([]string, 2)
	labels[0] = "key1=val1"
	labels[1] = "key2=val2"
	cfg := new(framework.ServiceConfig)
	applyKeyValSliceFlag(labels, func(configMap map[string]string) {
		if configMap != nil && len(configMap) != 0 {
			cfg.Labels = configMap
		}
	})
	assert.Equal(t, "val1", cfg.Labels["key1"], "Should contain key1")
	assert.Equal(t, "val2", cfg.Labels["key2"], "Should contain key2")
}
