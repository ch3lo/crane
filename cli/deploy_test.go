package cli

import (
        "testing"
        "github.com/stretchr/testify/assert"
	"github.com/latam-airlines/mesos-framework-factory"
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
        applyPorts([]string{"8080/tcp","443/tcp"}, cfg)
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
