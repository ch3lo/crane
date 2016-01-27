package cli

import (
	"flag"
	"github.com/codegangsta/cli"
	"github.com/latam-airlines/mesos-framework-factory"
	"github.com/stretchr/testify/assert"
	"testing"
)

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

func TestApplyConstraintsNil(t *testing.T) {
	err := applyConstraints(nil, "", nil)
	assert.Nil(t, err, "Should do nothing")
}

func TestApplyConstraintsBeta(t *testing.T) {
	cfg := new(framework.ServiceConfig)
	applyConstraints(nil, "beta4002", cfg)
	assert.Equal(t, "beta4002", cfg.Constraints["slave_name"], "Should contain beta")
}

func TestApplyConstraints(t *testing.T) {
	constraints := make([]string, 2)
	constraints[0] = "key1=val1"
	constraints[1] = "key2=val2"
	cfg := new(framework.ServiceConfig)
	applyConstraints(constraints, "beta4002", cfg)
	assert.Equal(t, "beta4002", cfg.Constraints["slave_name"], "Should contain beta")
	assert.Equal(t, "val1", cfg.Constraints["key1"], "Should contain key1")
	assert.Equal(t, "val2", cfg.Constraints["key2"], "Should contain key2")
}

func TestApplyConstraintsOnly(t *testing.T) {
	constraints := make([]string, 2)
	constraints[0] = "key1=val1"
	constraints[1] = "key2=val2"
	cfg := new(framework.ServiceConfig)
	applyConstraints(constraints, "", cfg)
	assert.Equal(t, "val1", cfg.Constraints["key1"], "Should contain key1")
	assert.Equal(t, "val2", cfg.Constraints["key2"], "Should contain key2")
}

func TestApplyConstraintsError(t *testing.T) {
	constraints := make([]string, 1)
	constraints[0] = "key2;val2"
	cfg := new(framework.ServiceConfig)
	err := applyConstraints(constraints, "", cfg)
	assert.NotNil(t, err, "Should fail")
}

func TestDeployBeforeError(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
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

	set = flag.NewFlagSet("test", 0)
	set.String("image", "someImage", "")
	set.String("tag", "someTag", "")
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
	set := flag.NewFlagSet("test", 0)
	set.String("image", "someImage", "")
	set.String("tag", "someTag", "")
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
	set := flag.NewFlagSet("test", 0)
	set.String("image", "someImage", "")
	set.String("tag", "someTag", "")
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
	set.String("framework", "marathon", "")

	endpointFlag := createEndpointSliceFlag()
	endpointFlag.Apply(set)
	ctx := cli.NewContext(nil, set, nil)

	deployCmd(ctx)
}

func TestCpuFlag(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("framework", "marathon", "some hint")
	set.String("image", "nginx", "some hint")
	set.String("tag", "latest", "some hint")
	set.Float64("cpu", 0.25, "usage")
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.Nil(t, err, "Should be fine")
}

func TestCpuFlagNegative(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("image", "nginx", "some hint")
	set.String("tag", "latest", "some hint")
	set.Float64("cpu", -2.1, "usage")
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.NotNil(t, err, "Should fail")
}

func TestCpuFlagMarathonWrongRange(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("framework", "marathon", "some hint")
	set.String("image", "nginx", "some hint")
	set.String("tag", "latest", "some hint")
	set.Float64("cpu", 1.1, "usage")
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.NotNil(t, err, "Should fail")
}

func TestCpuFlagSwarmWrongRange(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("framework", "swarm", "some hint")
	set.String("image", "nginx", "some hint")
	set.String("tag", "latest", "some hint")
	set.Float64("cpu", 1025, "usage")
	ctx := cli.NewContext(nil, set, nil)
	err := deployBefore(ctx)
	assert.NotNil(t, err, "Should fail")
}
