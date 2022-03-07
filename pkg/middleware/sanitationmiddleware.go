package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
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
						var output interface{}
						output = res.String()
						if fmt.Sprint(value) == output {
							output = value
						}
						val := sm.sanitizeType(table, column, output)
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

func (sm *SanitationMiddleware) sanitizeType(table *database.ReflectedTable, column *database.ReflectedColumn, value interface{}) interface{} {
	var newValue interface{}
	tables := sm.getArrayProperty("tables", "all")
	types := sm.getArrayProperty("types", "all")
	if (tables["all"] || tables[table.GetName()]) && (types["all"] || types[column.GetType()]) {
		if value == nil || value == "" {
			return value
		}
		newValue = value

		switch column.GetType() {
		case "integer", "bigint":
			switch t := value.(type) {
			case float64:
				newValue = int(math.Round(t))
			case string:
				if v, err := strconv.ParseFloat(strings.TrimSpace(t), 64); err == nil {
					newValue = int(math.Round(v))
				}
			}
		case "decimal":
			switch t := value.(type) {
			case float64:
				newValue = utils.NumberFormat(t, column.GetScale(), ".", "")
			case string:
				if v, err := strconv.ParseFloat(strings.TrimSpace(t), 64); err == nil {
					newValue = utils.NumberFormat(v, column.GetScale(), ".", "")
				}
			}
		case "float":
			if t, ok := value.(string); ok {
				if v, err := strconv.ParseFloat(strings.TrimSpace(t), 32); err == nil {
					newValue = v
				}
			}
		case "double":
			if t, ok := value.(string); ok {
				if v, err := strconv.ParseFloat(strings.TrimSpace(t), 64); err == nil {
					newValue = v
				}
			}
		case "boolean":
			switch t := value.(type) {
			case int, int8, int16, int32, int64:
				newValue = (t == 1)
			case float32, float64:
				newValue = (t == 1.0)
			case string:
				if v, err := strconv.ParseBool(strings.TrimSpace(t)); err == nil {
					newValue = v
				} else if t == "yes" || t == "on" { //Compat with php FILTER_VALIDATE_BOOLEAN filter
					newValue = true
				} else if t == "no" || t == "off" {
					newValue = false
				}
			}
		case "date":
			if m, ok := value.(string); ok {
				if v, err := strtotime.Parse(m, time.Now().Unix()); err == nil {
					t := time.Unix(v, 0).Local().UTC()
					newValue = fmt.Sprintf("%d-%02d-%02d", t.Year(), int(t.Month()), t.Day())
				}
			}
		case "time":
			if m, ok := value.(string); ok {
				if v, err := strtotime.Parse(m, time.Now().Unix()); err == nil {
					t := time.Unix(v, 0).Local().UTC()
					newValue = fmt.Sprintf("%02d:%02d:%02d", t.Hour(), int(t.Minute()), t.Second())
				}
			}
		case "timestamp":
			if m, ok := value.(string); ok {
				if v, err := strtotime.Parse(m, time.Now().Unix()); err == nil {
					t := time.Unix(v, 0).Local().UTC()
					newValue = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.Year(), int(t.Month()), t.Day(), t.Hour(), int(t.Minute()), t.Second())
				}
			}
		case "blob", "varbinary":
			if m, ok := value.(string); ok {
				// allow base64url format
				v := strings.ReplaceAll(strings.TrimSpace(m), `-`, `+`)
				newValue = strings.ReplaceAll(v, `_`, `/`)
			}
		case "clob", "varchar":
			newValue = value
		case "geometry":
			if m, ok := value.(string); ok {
				newValue = strings.TrimSpace(m)
			}
		}
		return newValue
	} else {
		return value
	}
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
