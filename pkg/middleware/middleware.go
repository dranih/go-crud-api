package middleware

import (
	"fmt"
	"net/http"
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
	for _, prop := range strings.Split(fmt.Sprint(gm.getProperty(key, defaut)), ",") {
		propMap[prop] = true
	}
	return propMap
}

func (gm *GenericMiddleware) getProperty(key, defaut string) interface{} {
	if val, exists := gm.Properties[key]; exists {
		return fmt.Sprint(val)
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
