package middleware

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"

	"github.com/Masterminds/sprig"
	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/dranih/go-crud-api/pkg/utils"

	"github.com/carmo-evan/strtotime"
)

type ValidationMiddleware struct {
	GenericMiddleware
	reflection *database.ReflectionService
}

func NewValidationMiddleware(responder controller.Responder, properties map[string]interface{}, reflection *database.ReflectionService) *ValidationMiddleware {
	return &ValidationMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}, reflection: reflection}
}

func (vm *ValidationMiddleware) callHandler(r *http.Request, w http.ResponseWriter, handler, operation string, table *database.ReflectedTable) bool {
	jsonMap, err := utils.GetBodyData(r)
	if err != nil || jsonMap == nil {
		return true
	}
	var records []map[string]interface{}
	switch v := jsonMap.(type) {
	case []map[string]interface{}:
		records = v
	case map[string]interface{}:
		records = append(records, v)
	default:
		return true
	}
	details := map[string]string{}

	tableName := table.GetName()
	if t, err := template.New("handler").Funcs(sprig.TxtFuncMap()).Parse(handler); err == nil {
		for i := range records {
			for columnName, value := range records[i] {
				if value != nil && table.HasColumn(columnName) {
					column := table.GetColumn(columnName)
					var res bytes.Buffer
					data := struct {
						Operation string
						TableName string
						Column    *database.ReflectedColumn
						Value     interface{}
						Context   []map[string]interface{}
					}{Operation: operation, TableName: tableName, Column: column, Value: value, Context: records}
					if err := t.Execute(&res, data); err == nil {
						var msg string
						allowed, _ := strconv.ParseBool(strings.TrimSpace(res.String()))
						if allowed {
							msg, allowed = vm.validateType(table, column, value)
						} else {
							msg = res.String()
						}
						if !allowed {
							details[columnName] = msg
						}
					} else {
						log.Printf("Error : could not execute template sanitation handler : %s", err.Error())
					}
				}
			}
		}
	} else {
		log.Printf("Error : could not parse template sanitation handler : %s", err.Error())
	}
	if len(details) > 0 {
		vm.Responder.Error(record.INPUT_VALIDATION_FAILED, tableName, w, details)
		return false
	}
	return true
}

func (vm *ValidationMiddleware) validateType(table *database.ReflectedTable, column *database.ReflectedColumn, value interface{}) (string, bool) {
	tables := vm.getArrayProperty("tables", "all")
	types := vm.getArrayProperty("types", "all")
	if (tables["all"] || tables[table.GetName()]) && (types["all"] || types[column.GetType()]) {
		if value == nil {
			if column.GetNullable() {
				return "", true
			} else {
				return "cannot be null", false
			}
		}
		if v, ok := value.(string); ok {
			switch column.GetType() {
			// check for whitespace
			case "varchar", "clob":
				break
			default:
				if len(strings.TrimSpace(v)) != len(v) {
					return "illegal whitespace", false
				}
			}
			// try to parse
			switch column.GetType() {
			case "integer", "bigint":
				if _, err := strconv.Atoi(v); err != nil {
					return "invalid integer", false
				}
			case "decimal":
				var whole, decimals string
				if strings.Index(v, ".") != -1 {
					a := strings.SplitN(strings.TrimLeft(v, "-"), ".", 2)
					whole = a[0]
					decimals = a[1]
				} else {
					whole = strings.TrimLeft(v, "-")
					decimals = ""
				}
				if _, err := strconv.Atoi(whole); err != nil && len(whole) > 0 {
					return "invalid decimal", false
				}
				if _, err := strconv.Atoi(decimals); err != nil && len(decimals) > 0 {
					return "invalid decimal", false
				}
				if len(whole) > column.GetPrecision()-column.GetScale() {
					return "decimal too large", false
				}
				if len(decimals) > column.GetScale() {
					return "decimal too precise", false
				}
			case "float":
				if _, err := strconv.ParseFloat(v, 32); err != nil {
					return "invalid float", false
				}
			case "double":
				if _, err := strconv.ParseFloat(v, 64); err != nil {
					return "invalid float", false
				}
			case "boolean":
				if _, err := strconv.ParseBool(v); err != nil {
					return "invalid boolean", false
				}
			case "date":
				re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
				if !re.MatchString(v) {
					return "invalid date", false
				}
			case "time":
				re := regexp.MustCompile(`\d{2}:\d{2}:\d{2}`)
				if !re.MatchString(v) {
					return "invalid time", false
				}
			case "timestamp":
				re := regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`)
				if !re.MatchString(v) {
					return "invalid timestamp", false
				}
			case "clob", "varchar":
				if column.HasLength() && utf8.RuneCountInString(v) > column.GetLength() {
					return "string too long", false
				}
				break
			case "blob", "varbinary":
				if val, err := base64.RawURLEncoding.DecodeString(v); err != nil {
					return "invalid base64", false
				} else if column.HasLength() && len(val) > column.GetLength() {
					return "string too long", false
				}
				break
			case "geometry":
				// no checks yet
				break
			}
		} else { // check non-string types
			switch column.GetType() {
			case "integer", "bigint":
				if _, ok := value.(int); !ok {
					if _, ok := value.(float64); !ok {
						return "invalid integer", false
					}
				}
			case "float", "double":
				if _, ok := value.(float64); !ok {
					return "invalid float", false
				}
			case "boolean":
				if _, ok := value.(bool); !ok {
					return "invalid boolean", false
				}
			case "date", "time", "timestamp":
				if _, ok := value.(time.Time); !ok {
					return "invalid date", false
				}
			default:
				return fmt.Sprintf("invalid %s", column.GetType()), false
			}
		}
	}
	return "", true
}

func (vm *ValidationMiddleware) sanitizeType(table *database.ReflectedTable, column *database.ReflectedColumn, value string) interface{} {
	var newValue interface{}
	tables := vm.getArrayProperty("tables", "all")
	types := vm.getArrayProperty("types", "all")
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

func (vm *ValidationMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		operation := utils.GetOperation(r)
		if operation == "create" || operation == "update" || operation == "increment" {
			tableName := utils.GetPathSegment(r, 2)
			if vm.reflection.HasTable(tableName) {
				handler := fmt.Sprint(vm.getProperty("handler", ""))
				if handler != "" {
					table := vm.reflection.GetTable(tableName)
					if ok := vm.callHandler(r, w, handler, operation, table); !ok {
						return
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
