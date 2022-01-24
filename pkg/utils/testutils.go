package utils

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
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

type Test struct {
	Name          string
	Method        string
	Uri           string
	Body          string
	Want          string
	WantRegex     string
	StatusCode    int
	Username      string
	Password      string
	AuthMethod    string
	Jar           http.CookieJar
	RequestHeader map[string]string
	WantHeader    map[string]string
}

func RunTests(t *testing.T, serverUrl string, tests []Test) {
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			request, err := http.NewRequest(tc.Method, serverUrl+tc.Uri, strings.NewReader(tc.Body))
			if err != nil {
				t.Fatal(err)
			}
			if tc.RequestHeader != nil {
				for header, value := range tc.RequestHeader {
					request.Header.Set(header, value)
				}
			}
			if tc.AuthMethod == "basicauth" && tc.Username != "" {
				request.SetBasicAuth(tc.Username, tc.Password)
			}
			var client *http.Client
			if tc.Jar != nil {
				client = &http.Client{
					Jar: tc.Jar,
				}
			} else {
				client = http.DefaultClient
			}
			resp, err := client.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.StatusCode {
				t.Errorf("Want status '%d', got '%d' at url '%s'", tc.StatusCode, resp.StatusCode, resp.Request.URL)
			}

			if tc.WantHeader != nil {
				for header, value := range tc.WantHeader {
					if gotValue := resp.Header.Get(header); gotValue != value {
						t.Errorf("Want header '%s : %s', got '%s : %s'", header, value, header, gotValue)
					}
				}
			}

			if b, err := io.ReadAll(resp.Body); err != nil {
				t.Errorf("Error reading response '%s'", err)
			} else {
				if tc.WantRegex != "" {
					re, _ := regexp.Compile(tc.WantRegex)
					if !re.Match(b) {
						t.Errorf("Regex '%s' not matching, got '%s'", tc.WantRegex, b)
					}
				} else if strings.TrimSpace(string(b)) != tc.Want {
					t.Errorf("Want '%s', got '%s'", tc.Want, strings.TrimSpace(string(b)))
				}
			}
		})
	}
}
