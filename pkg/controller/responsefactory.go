package controller

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type ResponseFactory struct{}

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
	content = bytes.Replace(content, []byte("\\u003c"), []byte("<"), -1)
	content = bytes.Replace(content, []byte("\\u003e"), []byte(">"), -1)
	content = bytes.Replace(content, []byte("\\u0026"), []byte("&"), -1)
	return rf.From(status, "application/json", content, w)
}

func (rf *ResponseFactory) From(status int, contentType string, content []byte, w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Content-Type", contentType+"; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(content); err != nil {
		log.Printf("ERROR : unable to write response : %s", err.Error())
	}
	return w
}

func (rf *ResponseFactory) FromStatus(status int, w http.ResponseWriter) http.ResponseWriter {
	w.WriteHeader(status)
	return w
}
