package testutils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetUrl(t *testing.T, url string, response interface{}) {
	res, err := http.Get(url + "/status/ping")
	if err != nil {
		t.Error(err, "Unable to complete Get request")
	} else {
		defer res.Body.Close()
		out, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Error(err, "Unable to read response data")
		} else {
			err = json.Unmarshal(out, &response)
			if err != nil {
				t.Error(err, "Unable to unmarshal response data")
			}
		}
	}
}
