package utils

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"reflect"
	"regexp"
	"runtime"
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
	SkipFor       map[string]bool
	Driver        string
	Server        string
	WantJson      string
}

func RunTests(t *testing.T, serverUrlHttps string, tests []Test) {
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			log.Printf("*************** Running test : %s ***************", tc.Name)
			//Skip test if requireGeo and sqlite
			if tc.SkipFor != nil && tc.Driver != "" && tc.SkipFor[tc.Driver] {
				log.Printf("Skipping test %s for driver %s", tc.Name, tc.Driver)
			} else {
				var url string
				if tc.Server == "" {
					url = serverUrlHttps + tc.Uri
				} else {
					url = tc.Server + tc.Uri
				}

				request, err := http.NewRequest(tc.Method, url, strings.NewReader(tc.Body))
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
				client := &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true,
						},
					},
				}
				if tc.Jar != nil {
					client.Jar = tc.Jar
				}
				//If we expect a http.StatusMovedPermanently, do not redirect
				if tc.StatusCode == http.StatusMovedPermanently {
					client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
						return http.ErrUseLastResponse
					}
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

				if b, err := ioutil.ReadAll(resp.Body); err != nil {
					t.Errorf("Error reading response '%s'", err)
				} else {
					if tc.WantRegex != "" {
						re, _ := regexp.Compile(tc.WantRegex)
						if !re.Match(b) {
							t.Errorf("Regex '%s' not matching, got '%s'", tc.WantRegex, b)
						}
					} else if tc.WantJson != "" {
						var j, j2 interface{}
						if err := json.Unmarshal([]byte(tc.WantJson), &j); err != nil {
							t.Errorf("Unmarshal error : %s", err.Error())
						}
						if err := json.Unmarshal(b, &j2); err != nil {
							t.Errorf("Unmarshal error : %s", err.Error())
						}
						if ok := JsonEqual(j, j2); !ok {
							t.Errorf("Want Json '%s'\nGot '%s'\nError : %s", tc.WantJson, strings.TrimSpace(string(b)), err)
						}
					} else if strings.TrimSpace(string(b)) != tc.Want {
						t.Errorf("Want '%s', got '%s'", tc.Want, strings.TrimSpace(string(b)))
					}
				}
			}
		})
	}
}

//For tests, if there is no GCA_CONFIG_FILE env var provided, we create a sqlite db and we use a default config file
func SelectConfig() string {
	if configFile := os.Getenv("GCA_CONFIG_FILE"); configFile != "" {
		return ""
	}
	//We create a sqlite db for the tests
	tmpFile, err := ioutil.TempFile(os.TempDir(), "gocrudtests-")
	if err != nil {
		log.Fatal("Cannot create temporary file", err)
	}
	filePath := tmpFile.Name()

	_, filename, _, _ := runtime.Caller(1)
	filepathSql := path.Join(path.Dir(filename), "../../test/sql/blog_sqlite.sql")

	dsn := fmt.Sprintf("%s?_fk=1&defer_fk=1&_auth&_auth_user=go-crud-api&_auth_pass=go-crud-api", filePath)
	if conn, err := sql.Open("sqlite3", dsn); err != nil {
		panic(fmt.Sprintf("Connection failed to database %s with error : %s", dsn, err))
	} else {
		if err := loadSqlFile(filepathSql, conn); err != nil {
			conn.Close()
			panic(err)
		}
		conn.Close()
	}
	filepathConfig := path.Join(path.Dir(filename), "../../test/yaml/gcaconfig_sqlite.yaml")
	os.Setenv("GCA_CONFIG_FILE", filepathConfig)
	return filePath
}

func IsServerStarted(url string) bool {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(request)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return true
	}
	return false
}

func loadSqlFile(sqlFile string, db *sql.DB) error {
	// Read file
	file, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		return err
	}

	// Execute all
	_, err = db.Exec(string(file))
	if err != nil {
		return err
	}
	return nil
}

// Equal checks equality between 2 Body-encoded data.
// https://github.com/emacampolo/gomparator/blob/master/json_util.go
func JsonEqual(vx, vy interface{}) bool {

	if reflect.TypeOf(vx) != reflect.TypeOf(vy) {
		return false
	}

	switch x := vx.(type) {
	case map[string]interface{}:
		y := vy.(map[string]interface{})

		if len(x) != len(y) {
			return false
		}

		for k, v := range x {
			val2 := y[k]

			if (v == nil) != (val2 == nil) {
				return false
			}

			if !JsonEqual(v, val2) {
				return false
			}
		}

		return true
	case []interface{}:
		y := vy.([]interface{})

		if len(x) != len(y) {
			return false
		}

		var matches int
		flagged := make([]bool, len(y))
		for _, v := range x {
			for i, v2 := range y {
				if JsonEqual(v, v2) && !flagged[i] {
					matches++
					flagged[i] = true

					break
				}
			}
		}

		return matches == len(x)
	default:
		return vx == vy
	}
}
