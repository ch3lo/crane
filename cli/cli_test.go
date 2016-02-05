package cli

import (
	"flag"
	"github.com/codegangsta/cli"
	"github.com/latam-airlines/crane/configuration"
	_ "github.com/latam-airlines/mesos-framework-factory/marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"reflect"
	"testing"
)

func TestCli(t *testing.T) {
	suite.Run(t, new(CliSuite))
}

type CliSuite struct {
	suite.Suite
	config    *configuration.Configuration
	globalSet *flag.FlagSet
}

func (suite *CliSuite) SetupTest() {
	suite.config = &configuration.Configuration{
		Clusters: map[string]configuration.Cluster{
			"local": {
				Framework: configuration.Framework{
					"marathon": configuration.Parameters{
						"address":        "http://localhost:8011",
						"tlsverify":      true,
						"tlscacert":      "ca-marathon.pem",
						"tlscert":        "cert-marathon.pem",
						"tlskey":         "key-marathon.pem",
						"deploy-timeout": 30,
					},
				},
			},
		},
	}

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.String("config", "/tmp/crane.yml", "config path")
	globalSet.String("log-level", "debug", "verbosidad")
	globalSet.String("log-formatter", "text", "formato")
	globalSet.String("log-output", "console", "output de logs")
	suite.globalSet = globalSet
}

func (suite *CliSuite) TestSetupApp() {
	globatCtx := cli.NewContext(nil, suite.globalSet, nil)
	err := setupApplication(globatCtx, func(configFile string) (*configuration.Configuration, error) {
		config := &configuration.Configuration{
			Logging: configuration.Loggging{
				Level:     "debug",
				Output:    "console",
				Formatter: "text",
			},
			Clusters: map[string]configuration.Cluster{
				"local": {
					Disabled: true,
					Framework: configuration.Framework{
						"marathon": configuration.Parameters{
							"address": "http://localhost:8011",
						},
					},
				},
				"remote": {
					Framework: configuration.Framework{
						"marathon": configuration.Parameters{
							"address":        "http://remote:8011",
							"deploy-timeout": 30,
						},
					},
				},
			},
		}

		return config, nil
	})
	assert.Nil(suite.T(), err, "Should return nil")
}

func (suite *CliSuite) TestInvalidLoglevel() {
	set := suite.globalSet
	set.Parse([]string{"--log-level=OtherLevel"})
	ctx := cli.NewContext(nil, set, nil)
	err := setupApplication(ctx, func(configFile string) (*configuration.Configuration, error) {
		return suite.config, nil
	})
	assert.NotNil(suite.T(), err, "Should return error")
}

func (suite *CliSuite) TestDisabledFramework() {
	ctx := cli.NewContext(nil, suite.globalSet, nil)
	err := setupApplication(ctx, func(configFile string) (*configuration.Configuration, error) {
		config := &configuration.Configuration{
			Logging: configuration.Loggging{
				Level:     "debug",
				Output:    "console",
				Formatter: "text",
			},
			Clusters: map[string]configuration.Cluster{
				"local": {
					Disabled: true,
					Framework: configuration.Framework{
						"marathon": configuration.Parameters{
							"address": "http://localhost:8011",
						},
					},
				},
				"remote": {
					Framework: configuration.Framework{
						"marathon": configuration.Parameters{
							"address":        "http://remote:8011",
							"deploy-timeout": 30,
						},
					},
				},
			},
		}

		return config, nil
	})
	assert.Nil(suite.T(), err, "Should return nil")

	v := reflect.ValueOf(stackManager).Elem()
	stacks := v.FieldByName("stacks")
	assert.Equal(suite.T(), 1, stacks.Len(), "Cli should instantiate one stack")
}

func TestReadConfiguration(t *testing.T) {
	res, _ := readConfiguration("../test/resources/crane.yml")
	assert.NotNil(t, res.Clusters["sjc"], "Cluster sjc should be set")
	_, err := readConfiguration("../test/resources/crane-not-there.yml")
	assert.NotNil(t, err, "Should throw error")
	_, err = readConfiguration("../test/resources/broken.yml")
	assert.NotNil(t, err, "Should throw error")
}
