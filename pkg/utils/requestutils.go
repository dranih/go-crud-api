package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("4d70ad3e4165e3a1dc158fdc0fc07dc9ae8c8c4ecbd7f8619debebcba5a3710d"))

func GetRequestParams(request *http.Request) map[string][]string {
	//params := map[string]string{}
	/*query := request.URL.RawQuery
	query = strings.Replace(strings.Replace(query, `=`, `%5B%5D=`, -1), `%5D%5B%5D=`, `%5D=`, -1)
	params, _ := url.ParseQuery(query)*/
	return request.URL.Query()
}

func GetSession(w http.ResponseWriter, request *http.Request) *sessions.Session {
	session, _ := store.Get(request, "session")
	// Save it before we write to the response/return from the handler.
	err := session.Save(request, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return session
}

//GetBodyData tries to get data from body request, as a urlencoded content type or as json by default
func GetBodyData(r *http.Request) (interface{}, error) {
	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		res := map[string]interface{}{}
		for key, val := range r.PostForm {
			if strings.HasSuffix(key, "__is_null") {
				res[strings.TrimSuffix(key, "__is_null")] = nil
			} else {
				res[key] = strings.Join(val, ",")
			}
		}
		return res, nil
	} else {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		var jsonMap interface{}
		err = json.Unmarshal(b, &jsonMap)
		if err != nil {
			return nil, err
		}
		return jsonMap, nil
	}
}

func GetBodyMapData(r *http.Request) (map[string]interface{}, error) {
	if res, err := GetBodyData(r); err != nil {
		return nil, err
	} else if resMap, ok := res.(map[string]interface{}); !ok {
		return nil, errors.New("unable to decode body")
	} else {
		return resMap, nil
	}
}

func GetPathSegment(r *http.Request, part int) string {
	path := r.URL.Path
	pathSegments := strings.Split(strings.TrimRight(path, "/"), "/")
	if part < 0 || part >= len(pathSegments) {
		return ""
	}
	return pathSegments[part]
}

func GetOperation(r *http.Request) string {
	method := r.Method
	path := GetPathSegment(r, 1)
	hasPk := false
	if GetPathSegment(r, 3) != "" {
		hasPk = true
	}
	switch path {
	case "openapi":
		return "document"
	case "columns":
		if method == "get" {
			return "reflect"
		} else {
			return "remodel"
		}
	case "geojson":
	case "records":
		switch method {
		case "POST":
			return "create"
		case "GET":
			if hasPk {
				return "read"
			} else {
				return "list"
			}
		case "PUT":
			return "update"
		case "DELETE":
			return "delete"
		case "PATCH":
			return "increment"
		}
	}
	return "unknown"
}

func GetTableNames(r *http.Request, allTableNames []string) []string {
	path := GetPathSegment(r, 1)
	tableName := GetPathSegment(r, 2)
	switch path {
	case "openapi":
		return allTableNames
	case "columns":
		if tableName != "" {
			return []string{tableName}
		} else {
			return allTableNames
		}
	case "records":
		return getJoinTables(tableName, GetRequestParams(r))
	}
	return allTableNames
}

func getJoinTables(tableName string, parameters map[string][]string) []string {
	uniqueTableNames := map[string]bool{}
	uniqueTableNames[tableName] = true
	if join, exists := parameters["join"]; exists {
		for _, parameter := range join {
			tableNames := strings.Split(strings.TrimSpace(parameter), ",")
			for _, tableNamef := range tableNames {
				uniqueTableNames[tableNamef] = true
			}
		}
	}
	var keys []string
	for key := range uniqueTableNames {
		keys = append(keys, key)
	}
	return keys
}
