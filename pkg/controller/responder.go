package controller

import "net/http"

type Responder interface {
	Error(errorCode int, argument interface{}, w http.ResponseWriter, details ...interface{}) http.ResponseWriter
	Success(result interface{}, w http.ResponseWriter) http.ResponseWriter
	Exception(err error, w http.ResponseWriter) http.ResponseWriter
	Multi(results interface{}, w http.ResponseWriter) http.ResponseWriter
}
