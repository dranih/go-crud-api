package controller

import (
	"net/http"
)

type Responder interface {
	Error(errorCode int, argument string, w http.ResponseWriter, details string) http.ResponseWriter
	Success(result interface{}, w http.ResponseWriter) http.ResponseWriter
	Exception(err error, w http.ResponseWriter) http.ResponseWriter
	Multi(results *[]map[string]interface{}, w http.ResponseWriter) http.ResponseWriter
}
