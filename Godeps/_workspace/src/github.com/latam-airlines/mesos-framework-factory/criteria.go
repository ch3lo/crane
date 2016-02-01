package framework

import "regexp"

// ServiceInformationCriteria esta interfaz sirve para crear filtros de ServiceInformation
type ServiceInformationCriteria interface {
	MeetCriteria(status []*ServiceInformation) []*ServiceInformation
}

// ImageNameAndImageTagRegexpCriteria filtra aquellos servicios cuyo nombre y tag de imagen no
// cumplen con la expresion regular FullImageNameRegexp
type ImageNameAndImageTagRegexpCriteria struct {
	FullImageNameRegexp *regexp.Regexp
}

// MeetCriteria aplica el filtro ImageNameAndImageTagRegexpCriteria y retorna un []*ServiceInformation
// con aquellos servicios que cumplen con el criterio
func (c *ImageNameAndImageTagRegexpCriteria) MeetCriteria(elements []*ServiceInformation) []*ServiceInformation {
	var filtered []*ServiceInformation
	for k, v := range elements {
		if c.FullImageNameRegexp.MatchString(v.FullImageName()) {
			filtered = append(filtered, elements[k])
		}
	}
	return filtered
}

// AndCriteria es un criterio que se puede aplicar para realizar un && sobre otros dos criterios
type AndCriteria struct {
	criteria      ServiceInformationCriteria
	otherCriteria ServiceInformationCriteria
}

// MeetCriteria aplica el filtro que tiene como objetivo realizar un && sobre dos criterios
func (c *AndCriteria) MeetCriteria(elements []*ServiceInformation) []*ServiceInformation {
	filtered := c.criteria.MeetCriteria(elements)
	return c.otherCriteria.MeetCriteria(filtered)
}

// OrCriteria es un criterio que se puede aplicar para realizar un || sobre otros dos criterios
type OrCriteria struct {
	criteria      ServiceInformationCriteria
	otherCriteria ServiceInformationCriteria
}

func getService(id string, services []*ServiceInformation) *ServiceInformation {
	for key, srv := range services {
		if id == srv.ID {
			return services[key]
		}
	}
	return nil
}

// MeetCriteria aplica el filtro que tiene como objetivo realizar un || sobre dos criterios
func (c *OrCriteria) MeetCriteria(elements []*ServiceInformation) []*ServiceInformation {
	filtered := c.criteria.MeetCriteria(elements)
	others := c.otherCriteria.MeetCriteria(elements)

	for key, v := range others {
		if getService(v.ID, filtered) == nil {
			filtered = append(filtered, others[key])
		}
	}

	return filtered
}
