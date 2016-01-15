package cli

import (
	"testing"
	"reflect"
	"os"
	"os/exec"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"io/ioutil"
	_ "github.com/latam-airlines/mesos-framework-factory/marathon"
)

func TestCliCmdDeploy(t *testing.T) {
	content, _ := ioutil.ReadFile("../test/resources/marathon_tasks_response.json")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "application/json")
                fmt.Fprintln(w, string(content))
        }))
        defer ts.Close()

	os.Args = append(os.Args, "--framework=marathon", "--endpoint="+ts.URL, "deploy", "--image=nginx", "--tag=latest")
	RunApp()
	v := reflect.ValueOf(stackManager).Elem()
	stacks := v.FieldByName("stacks")
	assert.Equal(t, 1, stacks.Len(), "Cli should instantiate at least one stack")
}

func TestInvalidFramework(t *testing.T) {
        os.Args = append(os.Args, "--framework=bla", "--endpoint=url", "deploy", "--image=nginx", "--tag=latest")
	if os.Getenv("BE_CRASHER") == "1" {
		RunApp()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestInvalidFramework")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	assert.NotNil(t, err, "Should return error")
}

