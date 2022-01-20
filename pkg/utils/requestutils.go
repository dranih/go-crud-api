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
			res[key] = strings.Join(val, ",")
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
