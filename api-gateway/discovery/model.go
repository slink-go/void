package discovery

import "fmt"

type Remote struct {
	Host   string `json:"host,omitempty"`
	Port   int    `json:"port,omitempty"`
	Status string `json:"status,omitempty"`
}

func (r Remote) String() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type Remotes map[string][]Remote

func (r *Remotes) Add(app string, remote Remote) {
	panic("implement me")
}
func (r *Remotes) Get(app string) []Remote {
	panic("implement me")
}
