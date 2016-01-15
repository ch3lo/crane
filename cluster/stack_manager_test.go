package cluster

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"reflect"
	"github.com/latam-airlines/mesos-framework-factory/factory"
)

func TestConstructor(t *testing.T) {
	sm := NewStackManager()
	assert.True(t, sm != nil, "Instance should be healthy")
	params := make(map[string]interface{})
	params["address"] = "http://localhost:8081/v2"
	helper, _ := factory.Create("marathon", params)
	sm.AppendStack(helper)
	v := reflect.ValueOf(sm).Elem()
	stacks := v.FieldByName("stacks")
	assert.Equal(t, 1, stacks.Len(), "Cli should instantiate at least one stack")
}
