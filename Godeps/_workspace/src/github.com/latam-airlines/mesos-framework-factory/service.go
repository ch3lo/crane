package framework

import (
	"regexp"
	"strings"
)

// InstanceStatus define el estado de una instancia
type InstanceStatus int

const (
	// InstanceUp Estado de una instancia que está OK
	InstanceUp InstanceStatus = 1 + iota
	// InstanceDown Estado de una instancia que esta caida
	InstanceDown
)

var statuses = [...]string{
	"Up",
	"Down",
}

func (s InstanceStatus) String() string {
	return statuses[s-1]
}

// InstancePortType Tipo de protocolo de un puerto
type InstancePortType string

const (
	// TCP Puerto de protocolo TCP
	TCP InstancePortType = "TCP"
	// UDP Puerto de protocolo UDP
	UDP InstancePortType = "UDP"
)

// NewInstancePortType retorna un InstancePortType basado en el string pasado como parametro
func NewInstancePortType(t string) InstancePortType {
	if strings.ToUpper(t) == "UDP" {
		return UDP
	}
	return TCP
}

// InstancePort estructura que encapsula la información relacionada a un puerto de una instancia
type InstancePort struct {
	Advertise string
	Internal  int64
	Publics   []int64
	Type      InstancePortType
}

// ServiceInformation define una estructura con la información básica de un servicio
// Esta estructura sirve para la comunicación con los consumidores de Frameworks
type ServiceInformation struct {
	ID        string
	ImageName string
	ImageTag  string
	Instances []*Instance
}

// FullImageName entrega el nombre completo de la imagen incluyendo el tag.
// Si el tag no existe se asume un tag por defecto latest
func (si ServiceInformation) FullImageName() string {
	if si.ImageTag == "" {
		return si.ImageName + ":latest"
	}
	return si.ImageName + ":" + si.ImageTag
}

// Instance encapsula la informacion de una instancia de servicio.
type Instance struct {
	ID            string
	Host          string
	ContainerName string
	Status        InstanceStatus
	Ports         map[string]InstancePort
}

func (i *Instance) Healthy() bool {
	return i.Status == InstanceUp
}

// ServiceConfig estructura que encapsula la informacion necesaria para crear un servicio
type ServiceConfig struct {
	ServiceID   string
	CPUShares   int
	Envs        []string
	ImageName   string
	Memory      int64
	Publish     []string
	Tag         string
	Constraints map[string]string
}

// Version retorna la version de un servicio
func (s *ServiceConfig) Version() string {
	rp := regexp.MustCompile("^([\\d\\.]+)-")
	result := rp.FindStringSubmatch(s.Tag)
	if result == nil {
		//util.Log.Fatalln("Formato de TAG invalido")
		return ""
	}
	return result[1]
}

func (s *ServiceConfig) String() string {
	return "" // TODO FIX
	//return fmt.Sprintf("ImageName: %s - Tag: %s - CpuShares: %d - Memory: %s - Publish: %#v - Envs: %s", s.ImageName, s.Tag, s.CPUShares, s.Memory, s.Publish, util.MaskEnv(s.Envs))
}

func (s *ServiceConfig) ConvertImageTagToServiceId() {
	if s.Tag == "" {
		s.ServiceID = s.convertImageString(s.ImageName)+"_LATEST"
	} else {
		s.ServiceID = s.convertImageString(s.ImageName)+"_" +s.convertTagString(s.Tag)
	}
}

func (s *ServiceConfig) convertImageString(token string) string {
        convertedString := strings.ToUpper(token)
        splitted := strings.Split(convertedString, "/")
        return s.replaceSpecialCharacters(splitted[len(splitted) - 1], "_")
}

func (s *ServiceConfig) convertTagString(token string) string {
        convertedString := strings.ToUpper(token)
        return s.replaceSpecialCharacters(convertedString, "_")
}


func (s *ServiceConfig) replaceSpecialCharacters(oldString, dest string) string {
        convertedString := strings.Replace(oldString, "-", dest, -1)
        convertedString = strings.Replace(convertedString, "/", dest, -1)
        convertedString = strings.Replace(convertedString, "%", dest, -1)
        return strings.Replace(convertedString, ".", "_", -1)
}
