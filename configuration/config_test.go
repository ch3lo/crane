package configuration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"gopkg.in/yaml.v2"
)

// configStruct is a canonical example configuration, which should map to configYaml
var configStruct = Configuration{
	Logging: Loggging{
		Level:     "info",
		Formatter: "text",
		Colored:   false,
		Output:    "crane.log",
	},
	Clusters: map[string]Cluster{
		"dal": {
			Framework: Framework{
				"swarm": Parameters{
					"address":   "1.1.1.1:2376",
					"tlsverify": true,
					"tlscacert": "ca-swarm.pem",
					"tlscert":   "cert-swarm.pem",
					"tlskey":    "key-swarm.pem",
					"authfile":  ".dockercfg",
				},
			},
		},
		"wdc": {
			Framework: Framework{
				"swarm": Parameters{
					"address":   "2.2.2.2:2376",
					"tlsverify": true,
					"tlscacert": "ca-swarm.pem",
					"tlscert":   "cert-swarm.pem",
					"tlskey":    "key-swarm.pem",
				},
			},
		},
		"sjc": {
			Disabled: true,
			Framework: Framework{
				"marathon": Parameters{
					"address":        "3.3.3.3:8081",
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

// configYaml document representing configStruct
var configYaml = `
logging:
  level: info
  formatter: text
  colored: false
  output: crane.log
cluster:
  dal:
    framework:
      swarm:
        address: 1.1.1.1:2376
        tlsverify: true
        tlscacert: ca-swarm.pem
        tlscert: cert-swarm.pem
        tlskey: key-swarm.pem
        authfile: .dockercfg
  wdc:
    framework:
      swarm:
        address: 2.2.2.2:2376
        tlsverify: true
        tlscacert: ca-swarm.pem
        tlscert: cert-swarm.pem
        tlskey: key-swarm.pem
  sjc:
    disabled: true
    framework:
      marathon:
        address: 3.3.3.3:8081
        tlsverify: true
        tlscacert: ca-marathon.pem
        tlscert: cert-marathon.pem
        tlskey: key-marathon.pem
        deploy-timeout: 30
`

func Test(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

type ConfigSuite struct {
	suite.Suite
	expectedConfig Configuration
}

func (suite *ConfigSuite) SetupTest() {
	os.Clearenv()
	suite.expectedConfig = configStruct
}

func (suite *ConfigSuite) TestMarshalRoundtrip() {
	configBytes, err := yaml.Marshal(suite.expectedConfig)
	assert.Nil(suite.T(), err)
	var config Configuration
	err = yaml.Unmarshal(configBytes, &config)
	assert.Nil(suite.T(), err)
	assert.True(suite.T(), assert.ObjectsAreEqual(config, suite.expectedConfig))
}

func (suite *ConfigSuite) TestParseSimple() {
	var config Configuration
	err := yaml.Unmarshal([]byte(configYaml), &config)
	assert.Nil(suite.T(), err)
	assert.True(suite.T(), assert.ObjectsAreEqual(config, suite.expectedConfig))
}
