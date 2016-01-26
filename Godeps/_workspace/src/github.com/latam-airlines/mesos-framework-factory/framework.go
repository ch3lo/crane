package framework

// Framework es una interfaz que debe implementar para la comunicacion con los Schedulers de Docker
// Para un ejemplo ir a swarm.Framework
type Framework interface {
	ID() string
	FindServiceInformation(ServiceInformationCriteria) ([]*ServiceInformation, error)
	UndeployInstance(string) error
	DeployService(ServiceConfig, int) (*ServiceInformation, error)
	DeleteService(string) error
}
