package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/utils"
)

type IpAddressMiddleware struct {
	GenericMiddleware
	reflection *database.ReflectionService
}

func NewIpAddressMiddleware(responder controller.Responder, properties map[string]interface{}, reflection *database.ReflectionService) *IpAddressMiddleware {
	return &IpAddressMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}, reflection: reflection}
}

func (iam *IpAddressMiddleware) callHandler(r *http.Request, operation, tableName string) *http.Request {
	jsonMap, err := utils.GetBodyData(r)
	if err != nil || jsonMap == nil {
		return r
	}
	var records []map[string]interface{}
	multi := false
	switch v := jsonMap.(type) {
	case []map[string]interface{}:
		records = v
		multi = true
	case map[string]interface{}:
		records = append(records, v)
	default:
		return r
	}
	table := iam.reflection.GetTable(tableName)
	touched := false
	if columnNames := fmt.Sprint(iam.getProperty("columns", "")); columnNames != "" {
		for _, columnName := range strings.Split(columnNames, ",") {
			if table.HasColumn(columnName) {
				for i := range records {
					if operation == "create" {
						records[i][columnName] = iam.getIpAddress(r)
					} else {
						delete(records[i], columnName)
					}
				}
				touched = true
			}
		}
	}
	if touched {
		var body []byte
		if multi {
			body, err = json.Marshal(records)
		} else {
			body, err = json.Marshal(records[0])
		}
		if err != nil {
			log.Printf("Error : could not marshal modified body to string : %s", err.Error())
			return r
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		r.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}
	return r
}

func (iam *IpAddressMiddleware) getIpAddress(r *http.Request) string {
	var ipAddress string
	reverseProxy := fmt.Sprint(iam.getProperty("reverseProxy", ""))
	if reverseProxy != "" {
		ipAddress = r.Header.Get("X-Forwarded-For")
	} else {
		if r.RemoteAddr != "" {
			ipAddress = r.RemoteAddr
		} else {
			ipAddress = "127.0.0.1"
		}
	}
	return ipAddress
}

func (iam *IpAddressMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		operation := utils.GetOperation(r)
		if operation == "create" || operation == "update" || operation == "increment" {
			tableNames := iam.getArrayProperty("tables", "")
			tableName := utils.GetPathSegment(r, 2)
			if len(tableNames) == 0 || tableNames[tableName] {
				if iam.reflection.HasTable(tableName) {
					r = iam.callHandler(r, operation, tableName)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
