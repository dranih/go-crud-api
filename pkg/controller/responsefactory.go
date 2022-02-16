package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/dranih/go-crud-api/pkg/utils"
)

type ResponseFactory struct {
	afterHandler *template.Template
}

func (rf *ResponseFactory) FromXml(status int, xml string, w http.ResponseWriter, r *http.Request) http.ResponseWriter {
	return rf.From(status, "text/xml", []byte(xml), w, r)
}

func (rf *ResponseFactory) FromCsv(status int, csv string, w http.ResponseWriter, r *http.Request) http.ResponseWriter {
	return rf.From(status, "text/csv", []byte(csv), w, r)
}

func (rf *ResponseFactory) FromHtml(status int, html string, w http.ResponseWriter, r *http.Request) http.ResponseWriter {
	return rf.From(status, "text/html", []byte(html), w, r)
}

// Should check marshalling error
func (rf *ResponseFactory) FromObject(status int, body interface{}, w http.ResponseWriter, r *http.Request) http.ResponseWriter {
	content, _ := json.Marshal(body)
	return rf.From(status, "application/json", content, w, r)
}

func (rf *ResponseFactory) From(status int, contentType string, content []byte, w http.ResponseWriter, r *http.Request) http.ResponseWriter {
	if rf.afterHandler != nil {
		rf.afterHandle(content, w, r)
	}
	w.Header().Set("Content-Type", contentType+"; charset=utf-8")
	w.WriteHeader(status)
	w.Write(content)
	return w
}

func (rf *ResponseFactory) FromStatus(status int, w http.ResponseWriter) http.ResponseWriter {
	w.WriteHeader(status)
	return w
}

func (rf *ResponseFactory) afterHandle(content []byte, w http.ResponseWriter, r *http.Request) {
	if session := utils.GetSession(w, r); session == nil {
		log.Printf("Error : could not load session")
		return
	} else if env, exists := session.Values["environment"]; !exists {
		log.Printf("Error : could not get environment from session")
		return
	} else if environment, ok := env.(map[string]interface{}); !ok {
		log.Printf("Error : could not read environment from session")
		return
	} else {
		headers := map[string]interface{}{}
		for key := range w.Header() {
			headers[key] = w.Header().Get(key)
			w.Header().Del("key")
		}
		data := struct {
			Operation   string
			TableName   string
			Content     []byte
			Headers     map[string]interface{}
			Environment map[string]interface{}
		}{Operation: fmt.Sprint(environment["operation"]), TableName: fmt.Sprint(environment["tablename"]), Content: content, Headers: headers, Environment: environment}
		var res bytes.Buffer
		if err := rf.afterHandler.Execute(&res, data); err != nil {
			log.Printf("Error : could not execute template afterHandler : %s", err.Error())
		} else {
			for key, val := range headers {
				w.Header().Set(key, fmt.Sprint(val))
			}
		}
		delete(session.Values, "environment")
	}
}
