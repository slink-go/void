package security

import (
	"encoding/json"
	"fmt"
	"github.com/slink-go/logging"
	"os"
	"strings"
)

type Option interface {
	apply(rp *responseParser) error
}
type mappingOption struct {
	value map[string]interface{}
}

func (o *mappingOption) apply(rp *responseParser) error {
	if o.value != nil {
		rp.dict = rp.flatten(o.value)
	}
	return nil
}

type filePathOption struct {
	value string
}

func (o *filePathOption) apply(rp *responseParser) error {
	buff, err := os.ReadFile(o.value)
	if err != nil {
		return fmt.Errorf("could not read auth response mapping file: %s", err)
	}
	d := make(map[string]interface{})
	if err := json.Unmarshal(buff, &d); err != nil {
		return fmt.Errorf("could not parse auth response mapping file: %s", err)
	}
	rp.dict = rp.flatten(d)
	return nil
}

func NewResponseParser(options ...Option) ResponseParser {
	parser := responseParser{
		dict:   make(map[string]interface{}),
		logger: logging.GetLogger("response-parser"),
	}
	for _, option := range options {
		if option != nil {
			option.apply(&parser)
		}
	}
	return &parser
}
func WithMapping(value map[string]interface{}) Option {
	return &mappingOption{
		value: value,
	}
}
func WithMappingFile(value string) Option {
	return &filePathOption{
		value: value,
	}
}

type ResponseParser interface {
	Parse(source map[string]interface{}) map[string]string
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
