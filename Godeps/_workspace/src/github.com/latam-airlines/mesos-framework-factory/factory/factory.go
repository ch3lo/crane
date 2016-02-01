package factory

import (
	"fmt"

	"github.com/latam-airlines/crane/Godeps/_workspace/src/github.com/latam-airlines/mesos-framework-factory"
	"github.com/latam-airlines/crane/Godeps/_workspace/src/github.com/latam-airlines/mesos-framework-factory/logger"
)

// frameworkFactories almacena una mapeo entre un identificador de framework y su constructor
var frameworkFactories = make(map[string]FrameworkFactory)

// FrameworkFactory es una interfaz para crear un Framework
// Cada Framework debe implementar esta interfaz y además llamar el metodo Register
// para registrar el constructor de la implementacion
type FrameworkFactory interface {
	Create(parameters map[string]interface{}) (framework.Framework, error)
}

// Register permite registrar a una implementación de Framework, de esta
// manera estara disponible mediante su ID para poder ser instanciado
func Register(name string, factory FrameworkFactory) {
	if factory == nil {
		logger.Instance().Fatal("Se debe pasar como argumento un SchedulerFactory")
	}
	_, registered := frameworkFactories[name]
	if registered {
		logger.Instance().Fatalf("SchedulerFactory %s ya está registrado", name)
	}

	frameworkFactories[name] = factory
}

// Create crea un Framework a partir de un ID y retorna la implementacion asociada a él.
// Si el Framework no estaba registrado se retornará un InvalidFramework
func Create(name string, parameters map[string]interface{}) (framework.Framework, error) {
	schedulerFactory, ok := frameworkFactories[name]
	if !ok {
		return nil, InvalidFramework{name}
	}
	return schedulerFactory.Create(parameters)
}

// InvalidFramework es una estructura de error utilizada cuando se instenta
// crear un Framework no registrado
type InvalidFramework struct {
	Name string
}

func (err InvalidFramework) Error() string {
	return fmt.Sprintf("Framework no esta registrado: %s", err.Name)
}
