package cli

import (
        "testing"
        "github.com/stretchr/testify/assert"
	"github.com/latam-airlines/mesos-framework-factory"
)

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
