package cluster

import "fmt"

// ClusterDisabled error generated if cluster is disabled
type ClusterDisabled struct {
	Name string
}

func (err ClusterDisabled) Error() string {
	return fmt.Sprintf("The cluster is not enabled: %s", err.Name)
}
