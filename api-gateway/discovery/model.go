package discovery

import (
	"fmt"
	"strings"
)

type Remote struct {
	App    string `json:"app,omitempty"`
	Scheme string `json:"scheme,omitempty"`
	Host   string `json:"host,omitempty"`
	Port   int    `json:"port,omitempty"`
	Status string `json:"status,omitempty"`
}

func (r Remote) String() string {
	if r.Scheme == "http" && r.Port == 80 {
		return fmt.Sprintf("%s://%s", r.Scheme, r.Host)
	} else if r.Scheme == "https" && r.Port == 443 {
		return fmt.Sprintf("%s://%s", r.Scheme, r.Host)
	}
	return fmt.Sprintf("%s://%s:%d", r.Scheme, r.Host, r.Port)
}
func (r Remote) Compare(other Remote) int {
	return strings.Compare(r.String(), other.String())
}

type Remotes struct {
	Data map[string][]Remote `json:"remotes,omitempty"`
}

func (r *Remotes) All() map[string][]Remote {
	return r.Data
}
func (r *Remotes) Add(app string, remote Remote) {
	if r.Data == nil {
		r.Data = make(map[string][]Remote)
	}
	if _, ok := r.Data[app]; !ok {
		r.Data[app] = make([]Remote, 0)
	}
	r.Data[app] = append(r.Data[app], remote)
}
func (r *Remotes) Get(app string) []Remote {
	if r.Data == nil {
		return []Remote{}
	}
	v, ok := r.Data[app]
	if !ok {
		return []Remote{}
	}
	return v
}
func (r *Remotes) List() []Remote {
	if r == nil || r.Data == nil {
		return []Remote{}
	}
	result := make([]Remote, 0)
	for _, v := range r.Data {
		result = append(result, v...)
	}
	return result
}
