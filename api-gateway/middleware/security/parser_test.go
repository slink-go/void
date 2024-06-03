package security

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	input = `{
	"id": 1,
	"value": 0.1,
    "obj": {
       	"id": "id-1",
		"ch": {
			"id": "test",
			"flag": true
		},
		"arr": ["a", "b", "c"]
    }
}
`

	dict = `{
	"id": "Ctx-Id",
	"value": "Ctx-Value",
    "obj": {
       	"id": "Ctx-Obj-Id",
		"ch": {
			"id": "Ctx-Obj-Ch-Id",
			"flag": "Ctx-Obj-Ch-Flag"
		},
		"arr": "Ctx-Obj-Arr"
    }
}
`
)

func TestJsonFlatten(t *testing.T) {

	v := make(map[string]interface{})
	err := json.Unmarshal([]byte(input), &v)
	if err != nil {
		t.Fatal(err)
	}

	parser := &responseParser{}

	vv := parser.flatten(v)

	assert.Contains(t, vv, "id")
	assert.Equal(t, 1, vv["id"])

	assert.Contains(t, vv, "value")
	assert.Equal(t, 0.1, vv["value"])

	assert.Contains(t, vv, "obj.id")
	assert.Equal(t, "id-1", vv["obj.id"])

	assert.Contains(t, vv, "obj.ch.id")
	assert.Equal(t, "test", vv["obj.ch.id"])

	assert.Contains(t, vv, "obj.ch.flag")
	assert.Equal(t, true, vv["obj.ch.flag"])

	assert.Contains(t, vv, "obj.arr")
	assert.Equal(t, vv["obj.arr"], "a,b,c")

}

func TestJsonParse(t *testing.T) {

	sourceMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(input), &sourceMap)
	if err != nil {
		t.Fatal(err)
	}

	dictMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(dict), &dictMap)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewResponseParser().WithMapping(dictMap)
	result := parser.Parse(sourceMap)

	assert.Contains(t, result, "Ctx-Id")
	assert.Equal(t, "1", result["Ctx-Id"])

	assert.Contains(t, result, "Ctx-Value")
	assert.Equal(t, "0.1", result["Ctx-Value"])

	assert.Contains(t, result, "Ctx-Obj-Id")
	assert.Equal(t, "id-1", result["Ctx-Obj-Id"])

	assert.Contains(t, result, "Ctx-Obj-Ch-Id")
	assert.Equal(t, "test", result["Ctx-Obj-Ch-Id"])

	assert.Contains(t, result, "Ctx-Obj-Ch-Flag")
	assert.Equal(t, "true", result["Ctx-Obj-Ch-Flag"])

	assert.Contains(t, result, "Ctx-Obj-Arr")
	assert.Equal(t, "a,b,c", result["Ctx-Obj-Arr"])

}
