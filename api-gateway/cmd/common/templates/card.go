package templates

type Card struct {
	Title     string   `json:"title"`
	Instances []string `json:"instances"`
	Color     string   `json:"color"`
}
