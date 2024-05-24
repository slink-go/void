package discovery

import (
	"fmt"
)

type Remote struct {
	App    string `json:"app,omitempty"`
	Scheme string `json:"scheme,omitempty"`
	Host   string `json:"host,omitempty"`
	Port   int    `json:"port,omitempty"`
	Status string `json:"status,omitempty"`
}

func (r Remote) String() string {
	//if strings.HasPrefix(r.Scheme, "http") {
	//	return fmt.Sprintf("%s://%s:%d", r.Scheme, r.Host, r.Port)
	//}
	if r.Scheme == "http" && r.Port == 80 {
		return fmt.Sprintf("%s://%s", r.Scheme, r.Host)
	} else if r.Scheme == "https" && r.Port == 443 {
		return fmt.Sprintf("%s://%s", r.Scheme, r.Host)
	}
	return fmt.Sprintf("%s://%s:%d", r.Scheme, r.Host, r.Port)
}

type Remotes struct {
	data map[string][]Remote
}

func (r *Remotes) All() map[string][]Remote {
	return r.data
}
func (r *Remotes) Add(app string, remote Remote) {
	if r.data == nil {
		r.data = make(map[string][]Remote)
	}
	if _, ok := r.data[app]; !ok {
		r.data[app] = make([]Remote, 0)
	}
	r.data[app] = append(r.data[app], remote)
}
func (r *Remotes) Get(app string) []Remote {
	if r.data == nil {
		return []Remote{}
	}
	v, ok := r.data[app]
	if !ok {
		return []Remote{}
	}
	return v
}
func (r *Remotes) List() []Remote {
	if r == nil || r.data == nil {
		return []Remote{}
	}
	result := make([]Remote, 0)
	for _, v := range r.data {
		result = append(result, v...)
	}
	return result
}
