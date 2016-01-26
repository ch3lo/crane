package cli

import (
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
