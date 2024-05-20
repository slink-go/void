package registry

import "fmt"

type Remote struct {
	Host   string `json:"host,omitempty"`
	Port   int    `json:"port,omitempty"`
	Status string `json:"status,omitempty"`
}

func (r Remote) String() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type ServiceRegistry interface {
	Get(applicationId string) (string, error)
}
