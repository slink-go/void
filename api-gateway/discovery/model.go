package discovery

import "fmt"

type Remote struct {
	App    string `json:"app,omitempty"`
	Host   string `json:"host,omitempty"`
	Port   int    `json:"port,omitempty"`
	Status string `json:"status,omitempty"`
}

func (r Remote) String() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type Remotes struct {
	data map[string][]Remote
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
