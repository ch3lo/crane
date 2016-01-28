package configuration

import (
	"fmt"
	"strings"
)

// Cluster estructura para la configuración de un cluster
type Cluster struct {
	Disabled  bool      `yaml:"disabled"`
	Framework Framework `yaml:"framework"`
}

// Configuration estructura para la configuracion global de Crane
type Configuration struct {
	Clusters map[string]Cluster `yaml:"cluster"`
}

// Parameters mapeo para manejar configuraciones de distintos tipos de datos
type Parameters map[string]interface{}

// Framework mapeo de un un Framework en base a su ID y sus parametros de configuración
type Framework map[string]Parameters

// UnmarshalYAML deserializa el mapa de frameworks
func (fw *Framework) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var fwParameters map[string]Parameters
	err := unmarshal(&fwParameters)
	if err == nil {
		if len(fwParameters) > 1 {
			types := make([]string, 0, len(fwParameters))
			for k := range fwParameters {
				types = append(types, k)
			}

			if len(types) > 1 {
				return fmt.Errorf("Se debe configurar sólo un Framework por cluster. Frameworks: %v", types)
			}
		}
		*fw = fwParameters
		return nil
	}

	var frameworkType string
	err = unmarshal(&frameworkType)
	if err == nil {
		*fw = Framework{frameworkType: Parameters{}}
		return nil
	}

	return err
}

// MarshalYAML serializa el mapa de frameworks
func (fw Framework) MarshalYAML() (interface{}, error) {
	if fw.Parameters() == nil {
		return fw.Type(), nil
	}
	return map[string]Parameters(fw), nil
}

// Parameters retorna los parametros de configuracion del Framework
func (fw Framework) Parameters() Parameters {
	return fw[fw.Type()]
}

// Type retorna el tipo de Framework
func (fw Framework) Type() string {
	var schedulerType []string

	for k := range fw {
		schedulerType = append(schedulerType, k)
	}
	if len(schedulerType) > 1 {
		panic("multiples Frameworks definidos: " + strings.Join(schedulerType, ", "))
	}
	if len(schedulerType) == 1 {
		return schedulerType[0]
	}
	return ""
}
