package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ResponseFactory struct {
}

const OK = 200
const MOVED_PERMANENTLY = 301
const FOUND = 302
const UNAUTHORIZED = 401
const FORBIDDEN = 403
const NOT_FOUND = 404
const METHOD_NOT_ALLOWED = 405
const CONFLICT = 409
const UNPROCESSABLE_ENTITY = 422
const FAILED_DEPENDENCY = 424
const INTERNAL_SERVER_ERROR = 500

func (rf *ResponseFactory) FromXml(status int, xml string, w http.ResponseWriter) http.ResponseWriter {
	return rf.From(status, "text/xml", []byte(xml), w)
}

func (rf *ResponseFactory) FromCsv(status int, csv string, w http.ResponseWriter) http.ResponseWriter {
	return rf.From(status, "text/csv", []byte(csv), w)
}

func (rf *ResponseFactory) FromHtml(status int, html string, w http.ResponseWriter) http.ResponseWriter {
	return rf.From(status, "text/html", []byte(html), w)
}

// Should check marshalling error
func (rf *ResponseFactory) FromObject(status int, body interface{}, w http.ResponseWriter) http.ResponseWriter {
	content, _ := json.Marshal(body)
	return rf.From(status, "application/json", content, w)
}

func (rf *ResponseFactory) From(status int, contentType string, content []byte, w http.ResponseWriter) http.ResponseWriter {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", contentType+"; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprint(len(content)))
	w.Write(content)
	return w
}

func (rf *ResponseFactory) FromStatus(status int, w http.ResponseWriter) http.ResponseWriter {
	w.WriteHeader(status)
	return w
}
