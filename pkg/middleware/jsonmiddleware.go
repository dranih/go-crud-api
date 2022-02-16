package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/utils"
)

type JsonMiddleware struct {
	GenericMiddleware
}

func NewJsonMiddleware(responder controller.Responder, properties map[string]interface{}) *JsonMiddleware {
	return &JsonMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}}
}

type jsonResponseWriter struct {
	http.ResponseWriter
	columnNames map[string]bool
}

func NewJsonResponseWriter(w http.ResponseWriter, columnNames map[string]bool) *jsonResponseWriter {
	return &jsonResponseWriter{w, columnNames}
}

func (jrw *jsonResponseWriter) Write(b []byte) (int, error) {
	var body interface{}
	if err := json.Unmarshal(b, &body); err != nil {
		log.Printf("Error : %s", err.Error())
		return jrw.ResponseWriter.Write(b)
	}
	conv := jrw.convert(body)
	if res, err := json.Marshal(conv); err != nil {
		log.Printf("Error : %s", err.Error())
		return jrw.ResponseWriter.Write(b)
	} else {
		return jrw.ResponseWriter.Write(res)
	}
}

func (jrw *jsonResponseWriter) convertValue(v string) interface{} {
	var res map[string]interface{}
	if err := json.Unmarshal([]byte(v), &res); err != nil {
		return v
	}
	return res
}

func (jrw *jsonResponseWriter) convert(obj interface{}) interface{} {
	if obj == nil {
		return obj
	}

	switch v := obj.(type) {
	case []interface{}:
		for key, val := range v {
			v[key] = jrw.convert(val)
		}
		obj = v
	case map[string]interface{}:
		for key, val := range v {
			if jrw.columnNames["all"] || jrw.columnNames[key] {
				v[key] = jrw.convert(val)
			}
		}
		obj = v
	case string:
		obj = jrw.convertValue(v)
	}
	return obj
}

func (jm *JsonMiddleware) convertJsonResponse(w http.ResponseWriter, columnNames map[string]bool) http.ResponseWriter {
	return NewJsonResponseWriter(w, columnNames)
}

func (jm *JsonMiddleware) convertJsonRequest(r *http.Request, columnNames map[string]bool) *http.Request {
	jsonMap, err := utils.GetBodyData(r)
	if err != nil || jsonMap == nil {
		return r
	}
	var res interface{}
	switch v := jsonMap.(type) {
	case []interface{}:
		for i, obj := range v {
			if val, ok := obj.(map[string]interface{}); ok {
				for key, value := range val {
					if columnNames["all"] || columnNames[key] {
						val[key] = jm.convertJsonRequestValue(value)
					}
				}
				v[i] = val
			}
		}
		res = v
	case []map[string]interface{}:
		for i, obj := range v {
			for key, value := range obj {
				if columnNames["all"] || columnNames[key] {
					v[i][key] = jm.convertJsonRequestValue(value)
				}
			}
		}
		res = v
	case map[string]interface{}:
		for key, value := range v {
			if columnNames["all"] || columnNames[key] {
				v[key] = jm.convertJsonRequestValue(value)
			}
		}
		res = v
	default:
		return r
	}

	var body []byte
	body, err = json.Marshal(res)
	if err != nil {
		log.Printf("Error : could not marshal modified body to string : %s", err.Error())
		return r
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(body))
	r.ContentLength = int64(len(body))
	r.Header.Set("Content-Type", "application/json; charset=UTF-8")
	return r
}

func (jm *JsonMiddleware) convertJsonRequestValue(value interface{}) interface{} {
	if _, ok := value.(map[string]interface{}); ok {
		if res, err := json.Marshal(value); err != nil {
			log.Printf("Error : %s", err.Error())
			return value
		} else {
			return string(res)
		}
	} else {
		return value
	}
}

func (jm *JsonMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		operation := utils.GetOperation(r)
		controllerPath := utils.GetPathSegment(r, 1)
		tableName := utils.GetPathSegment(r, 2)

		controllerPaths := jm.getArrayProperty("controllers", "records,geojson")
		tableNames := jm.getArrayProperty("tables", "all")
		columnNames := jm.getArrayProperty("columns", "all")

		if (controllerPaths["all"] || controllerPaths[controllerPath]) && (tableNames["all"] || tableNames[tableName]) {
			if operation == "create" || operation == "update" {
				r = jm.convertJsonRequest(r, columnNames)
			} else if operation == "read" || operation == "list" {
				w = jm.convertJsonResponse(w, columnNames)
			}
		}
		next.ServeHTTP(w, r)
	})
}
