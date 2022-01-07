package utils

import (
	"net/http"

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
