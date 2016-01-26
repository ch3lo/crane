package utils

import (
	"github.com/latam-airlines/mesos-framework-factory"
	"strings"
)

type Syslog struct {
	flags []*ConfigurerParameter
}

type ConfigurerParameter struct {
	Key   string
	Value string
}
type LogConfigurer interface {
	GetFlags() []*ConfigurerParameter
	AddFlag(key, value string)
}

func CreateSyslogConfigurer(cfg *framework.ServiceConfig) LogConfigurer {
	s := new(Syslog)

	imageElements := strings.Split(cfg.ImageName, "/")

        s.AddFlag("log-driver", "syslog")
        s.AddFlag("log-opt", "syslog-facility=local1")
        s.AddFlag("log-opt", "tag=\"docker/{{.ImageName}}|"+imageElements[len(imageElements) - 1]+"|{{.ID}}\"")
	return s
}

func (s *Syslog) AddFlag(key, value string) {
	if s.flags == nil {
		s.flags = make([]*ConfigurerParameter, 0)
	}
	param := new(ConfigurerParameter)
	param.Key = key
	param.Value = value
	s.flags = append(s.flags, param)
}

func (s *Syslog) GetFlags() []*ConfigurerParameter {
	return s.flags
}
