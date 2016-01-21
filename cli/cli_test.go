package cli

import (
	"testing"
	"reflect"
	"os"
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

	os.Args = []string {"crane", "--framework=marathon", "--endpoint="+ts.URL, "deploy", "--image=nginx", "--tag=latest"}
	RunApp()
	v := reflect.ValueOf(stackManager).Elem()
	stacks := v.FieldByName("stacks")
	assert.Equal(t, 1, stacks.Len(), "Cli should instantiate at least one stack")
}

func TestInvalidFramework(t *testing.T) {
	os.Args = []string {"crane", "--framework=bla", "--endpoint=ep", "deploy", "--image=nginx", "--tag=latest"}
	fmt.Println("Calling RunApp: ")
	err := RunApp()
	assert.NotNil(t, err, "Should return error")
}

func TestDeployTimeout(t *testing.T) {
	content, _ := ioutil.ReadFile("../test/resources/marathon_tasks_response.json")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "application/json")
            fmt.Fprintln(w, string(content))
        }))
    defer ts.Close()
	
    os.Args = []string {"crane", "--framework=marathon", "--endpoint="+ts.URL, "--deploy-timeout=20", "deploy", "--image=nginx", "--tag=latest"}
	err := RunApp()
	assert.Nil(t, err, "Error on RunApp should be nil")
}

func TestDeployTimeoutOptional(t *testing.T) {
	content, _ := ioutil.ReadFile("../test/resources/marathon_tasks_response.json")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "application/json")
            fmt.Fprintln(w, string(content))
        }))
    defer ts.Close()
	
    os.Args = []string {"crane", "--framework=marathon", "--endpoint="+ts.URL, "deploy", "--image=nginx", "--tag=latest"}
	err := RunApp()
	assert.Nil(t, err, "Error on RunApp should be nil")
}

func TestInvalidDeployTimeout(t *testing.T) {
	content, _ := ioutil.ReadFile("../test/resources/marathon_tasks_response.json")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Content-Type", "application/json")
            fmt.Fprintln(w, string(content))
        }))
    defer ts.Close()
	
    os.Args = []string {"crane", "--framework=marathon", "--endpoint="+ts.URL, "--deploy-timeout=NotaNumber", "deploy", "--image=nginx", "--tag=latest"}
	err := RunApp()
	assert.NotNil(t, err, "Error on RunApp should Not be nil")
	
}
