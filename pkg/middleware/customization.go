package middleware

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"text/template"

	sprig "github.com/Masterminds/sprig/v3"
	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/utils"
)

type CustomizationMiddleware struct {
	GenericMiddleware
	reflection *database.ReflectionService
}

func NewCustomizationMiddleware(responder controller.Responder, properties map[string]interface{}, reflection *database.ReflectionService) *CustomizationMiddleware {
	return &CustomizationMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}, reflection: reflection}
}

func (cm *CustomizationMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		operation := utils.GetOperation(r)
		tableName := utils.GetPathSegment(r, 2)
		env := map[string]interface{}{}

		beforeHandler := fmt.Sprint(cm.getProperty("beforeHandler", ""))
		if beforeHandler != "" {
			if t, err := template.New("beforeHandler").Funcs(sprig.TxtFuncMap()).Parse(beforeHandler); err == nil {
				data := struct {
					Operation   string
					TableName   string
					Request     *http.Request
					Environment map[string]interface{}
				}{Operation: operation, TableName: tableName, Request: r, Environment: env}
				var res bytes.Buffer
				if err := t.Execute(&res, data); err != nil {
					log.Printf("Error : could not execute template beforeHandler : %s", err.Error())
				}
			} else {
				log.Printf("Error : could not parse template beforeHandler : %s", err.Error())
			}
		}

		var templateHeader, templateBody *template.Template
		afterHandlerHeader := fmt.Sprint(cm.getProperty("afterHandlerHeader", ""))
		if afterHandlerHeader != "" {
			if afterHandlerHeader != "" {
				if t, err := template.New("afterHandlerHeader").Funcs(sprig.TxtFuncMap()).Parse(afterHandlerHeader); err == nil {
					templateHeader = t
				} else {
					log.Printf("Error : could not parse template afterHandlerHeader : %s", err.Error())
				}
			}
		}
		afterHandlerBody := fmt.Sprint(cm.getProperty("afterHandlerBody", ""))
		if afterHandlerBody != "" {
			if afterHandlerBody != "" {
				if t, err := template.New("afterHandlerBody").Funcs(sprig.TxtFuncMap()).Parse(afterHandlerBody); err == nil {
					templateBody = t
				} else {
					log.Printf("Error : could not parse template afterHandlerBody : %s", err.Error())
				}
			}
		}
		if templateHeader != nil || templateBody != nil {
			w = NewCustomResponseWriter(w, operation, tableName, env, templateHeader, templateBody)
		}

		next.ServeHTTP(w, r)
	})
}

type customResponseWriter struct {
	http.ResponseWriter
	operation      string
	tableName      string
	environment    map[string]interface{}
	templateHeader *template.Template
	templateBody   *template.Template
}

func NewCustomResponseWriter(w http.ResponseWriter, operation, tableName string, environment map[string]interface{}, templateHeader *template.Template, templateBody *template.Template) *customResponseWriter {
	return &customResponseWriter{w, operation, tableName, environment, templateHeader, templateBody}
}

func (crw *customResponseWriter) Write(b []byte) (int, error) {
	if crw.templateBody != nil {
		data := struct {
			Operation   string
			TableName   string
			Content     []byte
			Environment map[string]interface{}
		}{Operation: crw.operation, TableName: crw.tableName, Content: b, Environment: crw.environment}
		var res bytes.Buffer
		if err := crw.templateHeader.Execute(&res, data); err != nil {
			log.Printf("Error : could not execute template afterHandlerBody : %s", err.Error())
		}
	}
	return crw.ResponseWriter.Write(b)
}

func (crw *customResponseWriter) WriteHeader(statusCode int) {
	if crw.templateHeader != nil {
		headers := map[string]interface{}{}
		for key := range crw.Header() {
			headers[key] = crw.Header().Get(key)
			crw.Header().Del("key")
		}
		data := struct {
			Operation   string
			TableName   string
			Headers     map[string]interface{}
			Environment map[string]interface{}
		}{Operation: crw.operation, TableName: crw.tableName, Headers: headers, Environment: crw.environment}
		var res bytes.Buffer
		if err := crw.templateHeader.Execute(&res, data); err != nil {
			log.Printf("Error : could not execute template afterHandlerHeader : %s", err.Error())
		} else {
			for key, val := range headers {
				crw.Header().Set(key, fmt.Sprint(val))
			}
		}
	}
	crw.ResponseWriter.WriteHeader(statusCode)
}
