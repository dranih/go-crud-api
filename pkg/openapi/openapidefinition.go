package openapi

import (
	"encoding/json"
	"strings"
)

type OpenApiDefinition struct {
	root map[string]interface{}
}

func (oad *OpenApiDefinition) Set(path string, value interface{}) {
	parts := strings.Split(strings.Trim(path, `|`), `|`)
	current := &oad.root
	lastPtr := current
	var lastPart string
	for _, part := range parts {
		if _, exists := (*current)[part]; !exists {
			(*current)[part] = map[string]interface{}{}
		}
		lastPtr = current
		lastPart = part
		if lvl, ok := ((*current)[part].(map[string]interface{})); ok {
			current = &lvl
		}
	}
	if lastPart != "" {
		(*lastPtr)[lastPart] = value
	}
}

func (oad *OpenApiDefinition) Has(path string) bool {
	parts := strings.Split(strings.Trim(path, `|`), `|`)
	current := &oad.root
	for _, part := range parts {
		if _, exists := (*current)[part]; !exists {
			return false
		}
		if lvl, ok := ((*current)[part].(map[string]interface{})); ok {
			current = &lvl
		}
	}
	return true
}

// json marshaling for struct OpenApiDefinition
func (oad *OpenApiDefinition) MarshalJSON() ([]byte, error) {
	return json.Marshal(oad.root)
}
