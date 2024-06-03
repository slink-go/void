package security

import (
	"encoding/json"
	"fmt"
	"github.com/slink-go/logging"
	"os"
	"strings"
)

func NewResponseParser() ResponseParser {
	parser := responseParser{
		dict:   make(map[string]interface{}),
		logger: logging.GetLogger("response-parser"),
	}
	return &parser
}

type ResponseParser interface {
	Parse(source map[string]interface{}) map[string]string
	WithMapping(mapping map[string]interface{}) ResponseParser
	LoadMapping(filePath string) ResponseParser
}

type responseParser struct {
	dict   map[string]interface{}
	logger logging.Logger
}

func (p *responseParser) Parse(source map[string]interface{}) map[string]string {
	if p.dict == nil || len(p.dict) == 0 {
		p.logger.Info("auth response mapping not set")
		return make(map[string]string)
	}
	result := make(map[string]string)
	for K, V := range p.flatten(source) {
		v := fmt.Sprintf("%v", V)
		kk, kok := p.dict[K]
		if kok && v != "" {
			k := fmt.Sprintf("%v", kk)
			result[k] = v
		}
	}
	return result
}
func (p *responseParser) WithMapping(mapping map[string]interface{}) ResponseParser {
	p.dict = p.flatten(mapping)
	return p
}
func (p *responseParser) LoadMapping(filePath string) ResponseParser {
	buff, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Errorf("could not load auth response mapping file: %s", err)
		return p
	}
	d := make(map[string]interface{})
	if err := json.Unmarshal(buff, &d); err != nil {
		fmt.Errorf("could not parse auth response mapping file: %s", err)
	}
	p.dict = p.flatten(d)
	return p
}

func (p *responseParser) flatten(jsonMap map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	p.doFlatten("", jsonMap, result)
	return result
}
func (p *responseParser) doFlatten(keyPrefix string, source, target map[string]interface{}) {
	for k, v := range source {
		switch v.(type) {
		case map[string]interface{}:
			p.doFlatten(p.key(keyPrefix, k), v.(map[string]interface{}), target)
		case []interface{}:
			arr := v.([]interface{})
			var strs []string
			for _, vv := range arr {
				strs = append(strs, fmt.Sprintf("%v", vv))
			}
			target[p.key(keyPrefix, k)] = strings.Join(strs, ",")
		case string:
			target[p.key(keyPrefix, k)] = v
		case float64:
			vv := v.(float64)
			if p.isIntegral(vv) {
				target[p.key(keyPrefix, k)] = int(vv)
			} else {
				target[p.key(keyPrefix, k)] = v
			}
		case bool:
			target[p.key(keyPrefix, k)] = v
		}
	}
}
func (p *responseParser) key(keyPrefix, k string) string {
	if keyPrefix == "" {
		return k
	} else {
		return keyPrefix + "." + k
	}
}
func (p *responseParser) isIntegral(val float64) bool {
	return val == float64(int(val))
}
