package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dranih/go-crud-api/pkg/controller"
)

type Middleware interface {
	getArrayProperty(key, defaut string) map[string]bool
	getProperty(key, defaut string) string
	getMapProperty(key, defaut string) map[string]string
	Process(next http.Handler) http.Handler
}

type GenericMiddleware struct {
	Responder  controller.Responder
	Properties map[string]interface{}
}

func (gm *GenericMiddleware) getArrayProperty(key, defaut string) map[string]bool {
	propMap := map[string]bool{}
	properties := fmt.Sprint(gm.getProperty(key, defaut))
	if properties == "" {
		return nil
	}
	for _, prop := range strings.Split(properties, ",") {
		propMap[prop] = true
	}
	return propMap
}

func (gm *GenericMiddleware) getProperty(key string, defaut interface{}) interface{} {
	if val, exists := gm.Properties[key]; exists {
		return val
	}
	return defaut
}

func (gm *GenericMiddleware) getIntProperty(key string, defaut int) int {
	if val, exists := gm.Properties[key]; exists {
		switch v := val.(type) {
		case string:
			if a, err := strconv.Atoi(v); err == nil {
				return a
			}
		case int:
			return v
		}
	}
	return defaut
}

func (gm *GenericMiddleware) getInt64Property(key string, defaut int64) int64 {
	if val, exists := gm.Properties[key]; exists {
		switch v := val.(type) {
		case string:
			if a, err := strconv.ParseInt(v, 10, 64); err == nil {
				return a
			}
		case int:
			return int64(v)
		case int64:
			return v
		}
	}
	return defaut
}

func (gm *GenericMiddleware) getStringProperty(key string, defaut string) string {
	if val, exists := gm.Properties[key]; exists {
		switch v := val.(type) {
		case string:
			return v
		default:
			return fmt.Sprint(v)
		}
	}
	return defaut
}

func (gm *GenericMiddleware) getMapProperty(key, defaut string) map[string]string {
	pairs := gm.getArrayProperty(key, defaut)
	result := map[string]string{}
	for pair := range pairs {
		if strings.Contains(pair, ":") {
			val := strings.SplitN(pair, ":", 2)
			result[val[1]] = val[2]
		} else {
			result[pair] = ""
		}
	}
	return result
}
