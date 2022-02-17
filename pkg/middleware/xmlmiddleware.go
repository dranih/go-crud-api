package middleware

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/utils"

	mxj "github.com/clbanning/mxj/v2"
)

type XmlMiddleware struct {
	GenericMiddleware
}

func NewXmlMiddleware(responder controller.Responder, properties map[string]interface{}) *XmlMiddleware {
	return &XmlMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}}
}

func (xm *XmlMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := utils.GetRequestParams(r)
		format := params.Get("format")
		if format == "xml" {
			types := xm.getArrayProperty("types", "null,array")
			w = NewXmlResponseWriter(w, types)
		}
		next.ServeHTTP(w, r)
	})
}

type xmlResponseWriter struct {
	http.ResponseWriter
	types map[string]bool
}

func NewXmlResponseWriter(w http.ResponseWriter, types map[string]bool) *xmlResponseWriter {
	return &xmlResponseWriter{w, types}
}

func (xm *xmlResponseWriter) Write(b []byte) (int, error) {
	if len(b) == 1 {
		res := fmt.Sprintf("<root>%s</root>", b)
		return xm.ResponseWriter.Write([]byte(res))
	}
	dataMap, err := mxj.NewMapJson(b)
	if err != nil {
		log.Printf("Error : %s", err.Error())
		return xm.ResponseWriter.Write(b)
	}
	if xmlData, err := dataMap.Xml("root"); err != nil {
		log.Printf("Error : %s", err.Error())
		return xm.ResponseWriter.Write(b)
	} else {
		return xm.ResponseWriter.Write(xmlData)
	}
}

func (xm *xmlResponseWriter) WriteHeader(statusCode int) {
	xm.ResponseWriter.Header().Set("Content-Type", "text/xml; charset=utf-8")
	xm.ResponseWriter.WriteHeader(statusCode)
}
