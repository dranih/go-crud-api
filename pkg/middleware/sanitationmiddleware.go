package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	sprig "github.com/Masterminds/sprig/v3"
	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/utils"

	"github.com/carmo-evan/strtotime"
)

type SanitationMiddleware struct {
	GenericMiddleware
	reflection *database.ReflectionService
}

func NewSanitationMiddleware(responder controller.Responder, properties map[string]interface{}, reflection *database.ReflectionService) *SanitationMiddleware {
	return &SanitationMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}, reflection: reflection}
}

func (sm *SanitationMiddleware) callHandler(r *http.Request, handler, operation string, table *database.ReflectedTable) *http.Request {
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
	tableName := table.GetName()
	if t, err := template.New("handler").Funcs(sprig.TxtFuncMap()).Parse(handler); err == nil {
		for i := range records {
			for columnName, value := range records[i] {
				//Skip if value is map[string]interface{}
				_, skip := value.(map[string]interface{})
				if value != nil && !skip && table.HasColumn(columnName) {
					column := table.GetColumn(columnName)
					var res bytes.Buffer
					data := struct {
						Operation string
						TableName string
						Column    string
						Value     interface{}
					}{Operation: operation, TableName: tableName, Column: columnName, Value: value}
					if err := t.Execute(&res, data); err == nil {
						val := sm.sanitizeType(table, column, res.String())
						records[i][columnName] = val
					} else {
						log.Printf("Error : could not execute template sanitation handler : %s", err.Error())
					}
				}
			}
		}
	} else {
		log.Printf("Error : could not parse template sanitation handler : %s", err.Error())
	}

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
	return r
}

func (sm *SanitationMiddleware) sanitizeType(table *database.ReflectedTable, column *database.ReflectedColumn, value string) interface{} {
	var newValue interface{}
	tables := sm.getArrayProperty("tables", "all")
	types := sm.getArrayProperty("types", "all")
	if (tables["all"] || tables[table.GetName()]) && (types["all"] || types[column.GetType()]) {
		if value == "" {
			return value
		}
		newValue = value
		switch column.GetType() {
		case "integer", "bigint":
			if v, err := strconv.ParseFloat(value, 0); err == nil {
				newValue = v
			}
		case "decimal":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				newValue = strconv.FormatFloat(v, 'g', column.GetScale(), 64)
			}
		case "float":
			if v, err := strconv.ParseFloat(value, 32); err == nil {
				newValue = v
			}
		case "double":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				newValue = v
			}
		case "boolean":
			if v, err := strconv.ParseBool(value); err == nil {
				newValue = v
			}
		case "date":
			if v, err := strtotime.Parse(value, time.Now().Unix()); err != nil {
				t := time.Unix(v, 0)
				newValue = fmt.Sprintf("%d-%02d-%02d", t.Year(), int(t.Month()), t.Day())
			}
		case "time":
			if v, err := strtotime.Parse(value, time.Now().Unix()); err != nil {
				t := time.Unix(v, 0)
				newValue = fmt.Sprintf("%02d:%02d:%02d", t.Hour(), int(t.Minute()), t.Second())
			}
		case "timestamp":
			if v, err := strtotime.Parse(value, time.Now().Unix()); err != nil {
				t := time.Unix(v, 0)
				newValue = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.Year(), int(t.Month()), t.Day(), t.Hour(), int(t.Minute()), t.Second())
			}
		case "blob", "varbinary":
			// allow base64url format
			v := strings.ReplaceAll(strings.TrimSpace(value), `-`, `+`)
			newValue = strings.ReplaceAll(v, `_`, `/`)
		case "clob", "varchar":
			newValue = value
		case "geometry":
			newValue = strings.TrimSpace(value)
		}
	}
	return newValue
}

func (sm *SanitationMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		operation := utils.GetOperation(r)
		if operation == "create" || operation == "update" || operation == "increment" {
			tableName := utils.GetPathSegment(r, 2)
			if sm.reflection.HasTable(tableName) {
				handler := fmt.Sprint(sm.getProperty("handler", ""))
				if handler != "" {
					table := sm.reflection.GetTable(tableName)
					r = sm.callHandler(r, handler, operation, table)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
