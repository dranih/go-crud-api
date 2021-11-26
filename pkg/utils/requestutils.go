package utils

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func GetRequestParams(request *http.Request) map[string][]string {
	//params := map[string]string{}
	/*query := request.URL.RawQuery
	query = strings.Replace(strings.Replace(query, `=`, `%5B%5D=`, -1), `%5D%5B%5D=`, `%5D=`, -1)
	params, _ := url.ParseQuery(query)*/
	return request.URL.Query()
}

func SetSession(w http.ResponseWriter, request *http.Request) *sessions.Session {
	var store = sessions.NewCookieStore([]byte("toto"))
	session, _ := store.Get(request, "session")
	// Set some session values.
	session.Values["foo"] = "bar"
	session.Values[42] = 43
	// Save it before we write to the response/return from the handler.
	err := session.Save(request, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return session
}
