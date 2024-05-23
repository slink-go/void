package templates

import (
	"github.com/slink-go/api-gateway/discovery"
	"slices"
)

var colorIndex = 0
var colors = []string{
	"lightsalmon",
	"lightgreen",
	"lightblue",
	"lightgoldenrodyellow",
	"lightskyblue",
	"lightgrey",
	"lightpink",
	"lightsalmon",
	"lightsalt",
}

func getColor() string {
	if colorIndex >= len(colors) {
		colorIndex = 0
	}
	color := colors[colorIndex]
	colorIndex++
	return color
}

func Cards(remotes []discovery.Remote) []Card {
	var result []Card

	services := make(map[string][]discovery.Remote)
	for _, r := range remotes {
		if _, ok := services[r.App]; !ok {
			services[r.App] = []discovery.Remote{}
		}
		services[r.App] = append(services[r.App], r)
	}
	keys := make([]string, 0, len(services))
	for k, _ := range services {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		instances := make([]string, 0)
		for _, instance := range services[k] {
			instances = append(instances, instance.String())
		}
		result = append(result, Card{
			Title:     k,
			Instances: instances,
			Color:     getColor(),
		})
	}
	return result
}
