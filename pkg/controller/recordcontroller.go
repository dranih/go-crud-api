package controller

import "net/http"

type RecordController struct {
	service   string
	responder http.ResponseWriter
}
