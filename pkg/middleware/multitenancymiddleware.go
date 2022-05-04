package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/utils"

	sprig "github.com/Masterminds/sprig/v3"

	"text/template"
)

type MultiTenancy struct {
	GenericMiddleware
	reflection *database.ReflectionService
}

func NewMultiTenancyMiddleware(responder controller.Responder, properties map[string]interface{}, reflection *database.ReflectionService) *MultiTenancy {
	return &MultiTenancy{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}, reflection: reflection}
}

func (mt *MultiTenancy) getCondition(tableName string, pairs map[string]string) interface{ database.Condition } {
	var condition interface{ database.Condition }
	condition = database.NewNoCondition()
	table := mt.reflection.GetTable(tableName)
	for k, v := range pairs {
		condition = condition.And(database.NewColumnCondition(table.GetColumn(k), "eq", v)).(interface{ database.Condition })
	}
	return condition
}

func (mt *MultiTenancy) getPairs(handler, operation, tableName string) map[string]string {
	result := map[string]string{}
	if t, err := template.New("handler").Funcs(sprig.TxtFuncMap()).Parse(handler); err == nil {
		var res bytes.Buffer
		data := struct {
			Operation string
			TableName string
		}{Operation: operation, TableName: tableName}
		if err := t.Execute(&res, data); err == nil {
			//We expect a map[string]interface{}
			var data map[string]interface{}
			if err := json.Unmarshal(res.Bytes(), &data); err == nil {
				table := mt.reflection.GetTable(tableName)
				for k, v := range data {
					if table.HasColumn(k) {
						result[k] = fmt.Sprint(v)
					}
				}
			} else {
				log.Printf("Error : could not unmarshal json from multitenancy handler : %s", err.Error())
			}
		} else {
			log.Printf("Error : could not execute template multitenancy handler : %s", err.Error())
		}
	} else {
		log.Printf("Error : could not parse template multitenancy handler : %s", err.Error())
	}
	return result
}

func (mt *MultiTenancy) handleRecord(r *http.Request, operation string, pairs map[string]string) *http.Request {
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
	for i := range records {
		for column, value := range pairs {
			if operation == "create" {
				records[i][column] = value
			} else {
				delete(records[i], column)
			}
		}
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

func (mt *MultiTenancy) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := fmt.Sprint(mt.getProperty("handler", ""))
		if handler != "" {
			path := utils.GetPathSegment(r, 1)
			if path == "records" {
				operation := utils.GetOperation(r)
				tableNames := utils.GetTableNames(r, mt.reflection.GetTableNames())
				for i, tableName := range tableNames {
					if !mt.reflection.HasTable(tableName) {
						continue
					}
					if pairs := mt.getPairs(handler, operation, tableName); len(pairs) > 0 {
						if i == 0 {
							if operation == "create" || operation == "update" || operation == "increment" {
								r = mt.handleRecord(r, operation, pairs)
							}
						}
						condition := mt.getCondition(tableName, pairs)
						utils.VStore.Set(fmt.Sprintf("multiTenancy.conditions.%s", tableName), condition)
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
