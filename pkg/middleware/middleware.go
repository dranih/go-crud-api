package middleware

import (
	"net/http"
	"strings"
)

type Middleware interface {
	getArrayProperty(key, defaut string) []string
	getProperty(key, defaut string) string
	getMapProperty(key, defaut string) map[string]string
	Process(next http.Handler) http.Handler
}

type GenericMiddleware struct {
	properties map[string]string
}

func (gm *GenericMiddleware) getArrayProperty(key, defaut string) []string {
	return strings.Split(gm.getProperty(key, defaut), ",")
}

func (gm *GenericMiddleware) getProperty(key, defaut string) string {
	if val, exists := gm.properties[key]; exists {
		return val
	}
	return defaut
}

func (gm *GenericMiddleware) getMapProperty(key, defaut string) map[string]string {
	pairs := gm.getArrayProperty(key, defaut)
	result := map[string]string{}
	for _, pair := range pairs {
		if strings.Contains(pair, ":") {
			val := strings.SplitN(pair, ":", 2)
			result[val[1]] = val[2]
		} else {
			result[pair] = ""
		}
	}
	return result
}
