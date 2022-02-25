package apiserver

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/dranih/go-crud-api/pkg/utils"
)

// Global API tests for records
// To check compatibility with php-crud-api
func TestColumnsApi(t *testing.T) {
	utils.SelectConfig()
	config := ReadConfig()
	config.Init()
	serverStarted := new(sync.WaitGroup)
	serverStarted.Add(1)
	api := NewApi(config)
	go api.Handle(serverStarted)
	//Waiting http server to start
	serverStarted.Wait()
	serverUrlHttps := fmt.Sprintf("https://%s:%d", config.Server.Address, config.Server.HttpsPort)

	//https://ieftimov.com/post/testing-in-go-testing-http-servers/
	//https://stackoverflow.com/questions/42474259/golang-how-to-live-test-an-http-server
	tt := []utils.Test{
		//Sqlite : no geometry and bigint ids
		{
			Name:   "001_get_database_A",
			Method: http.MethodGet,
			Uri:    "/columns",
			Body:   ``,
			//WantJson: `{"tables":[{"name":"barcodes","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"product_id","type":"integer","fk":"products"},{"name":"hex","type":"varchar","length":255},{"name":"bin","type":"blob"},{"name":"ip_address","type":"varchar","length":15,"nullable":true}]},{"name":"categories","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"name","type":"varchar","length":255},{"name":"icon","type":"blob","nullable":true}]},{"name":"comments","type":"table","columns":[{"name":"id","type":"bigint","pk":true},{"name":"post_id","type":"integer","fk":"posts"},{"name":"message","type":"varchar","length":255},{"name":"category_id","type":"integer","fk":"categories"}]},{"name":"countries","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"name","type":"varchar","length":255},{"name":"shape","type":"geometry"}]},{"name":"events","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"name","type":"varchar","length":255},{"name":"datetime","type":"timestamp","nullable":true},{"name":"visitors","type":"bigint","nullable":true}]},{"name":"kunsthåndværk","type":"table","columns":[{"name":"id","type":"varchar","length":36,"pk":true},{"name":"Umlauts ä_ö_ü-COUNT","type":"integer"},{"name":"user_id","type":"integer","fk":"users"},{"name":"invisible_id","type":"varchar","length":36,"nullable":true,"fk":"invisibles"}]},{"name":"nopk","type":"table","columns":[{"name":"id","type":"varchar","length":36}]},{"name":"post_tags","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"post_id","type":"integer","fk":"posts"},{"name":"tag_id","type":"integer","fk":"tags"}]},{"name":"posts","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"user_id","type":"integer","fk":"users"},{"name":"category_id","type":"integer","fk":"categories"},{"name":"content","type":"varchar","length":255}]},{"name":"products","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"name","type":"varchar","length":255},{"name":"price","type":"decimal","precision":10,"scale":2},{"name":"properties","type":"clob"},{"name":"created_at","type":"timestamp"},{"name":"deleted_at","type":"timestamp","nullable":true}]},{"name":"tag_usage","type":"view","columns":[{"name":"id","type":"integer","pk":true},{"name":"name","type":"varchar","length":255},{"name":"count","type":"bigint"}]},{"name":"tags","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"name","type":"varchar","length":255},{"name":"is_important","type":"boolean"}]},{"name":"users","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"username","type":"varchar","length":255},{"name":"password","type":"varchar","length":255},{"name":"api_key","type":"varchar","length":255,"nullable":true},{"name":"location","type":"geometry","nullable":true}]}]}`,
			//Sqlite : no geometry and bigint ids
			WantJson:   `{"tables":[{"columns":[{"name":"bin","type":"blob"},{"length":255,"name":"hex","type":"varchar"},{"name":"id","pk":true,"type":"integer"},{"length":15,"name":"ip_address","nullable":true,"type":"varchar"},{"fk":"products","name":"product_id","type":"integer"}],"name":"barcodes","type":"table"},{"columns":[{"name":"icon","nullable":true,"type":"blob"},{"name":"id","pk":true,"type":"integer"},{"length":255,"name":"name","type":"varchar"}],"name":"categories","type":"table"},{"columns":[{"fk":"categories","name":"category_id","type":"integer"},{"name":"id","pk":true,"type":"integer"},{"length":255,"name":"message","type":"varchar"},{"fk":"posts","name":"post_id","type":"integer"}],"name":"comments","type":"table"},{"columns":[{"name":"id","pk":true,"type":"integer"},{"length":255,"name":"name","type":"varchar"},{"name":"shape","type":"clob"}],"name":"countries","type":"table"},{"columns":[{"name":"datetime","nullable":true,"type":"timestamp"},{"name":"id","pk":true,"type":"integer"},{"length":255,"name":"name","type":"varchar"},{"name":"visitors","nullable":true,"type":"bigint"}],"name":"events","type":"table"},{"columns":[{"name":"Umlauts ä_ö_ü-COUNT","type":"integer"},{"length":36,"name":"id","pk":true,"type":"varchar"},{"fk":"invisibles","length":36,"name":"invisible_id","nullable":true,"type":"varchar"},{"fk":"users","name":"user_id","type":"integer"}],"name":"kunsthåndværk","type":"table"},{"columns":[{"length":36,"name":"id","type":"varchar"}],"name":"nopk","type":"table"},{"columns":[{"name":"id","pk":true,"type":"integer"},{"fk":"posts","name":"post_id","type":"integer"},{"fk":"tags","name":"tag_id","type":"integer"}],"name":"post_tags","type":"table"},{"columns":[{"fk":"categories","name":"category_id","type":"integer"},{"length":255,"name":"content","type":"varchar"},{"name":"id","pk":true,"type":"integer"},{"fk":"users","name":"user_id","type":"integer"}],"name":"posts","type":"table"},{"columns":[{"name":"created_at","type":"timestamp"},{"name":"deleted_at","nullable":true,"type":"timestamp"},{"name":"id","pk":true,"type":"integer"},{"length":255,"name":"name","type":"varchar"},{"name":"price","precision":10,"scale":2,"type":"decimal"},{"name":"properties","type":"clob"}],"name":"products","type":"table"},{"columns":[{"name":"count","type":"clob"},{"name":"id","pk":true,"type":"integer"},{"length":255,"name":"name","type":"varchar"}],"name":"tag_usage","type":"view"},{"columns":[{"name":"id","pk":true,"type":"integer"},{"name":"is_important","type":"boolean"},{"length":255,"name":"name","type":"varchar"}],"name":"tags","type":"table"},{"columns":[{"length":255,"name":"api_key","nullable":true,"type":"varchar"},{"name":"id","pk":true,"type":"integer"},{"name":"location","nullable":true,"type":"clob"},{"length":255,"name":"password","type":"varchar"},{"length":255,"name":"username","type":"varchar"}],"name":"users","type":"table"}]}`,
			StatusCode: http.StatusOK,
			Driver:     config.Api.Driver,
			SkipFor:    map[string]bool{"mysql": true, "pgsql": true, "sqlsrv": true},
		},
		{
			Name:       "001_get_database_B",
			Method:     http.MethodGet,
			Uri:        "/columns",
			Body:       ``,
			WantJson:   `{"tables":[{"name":"barcodes","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"product_id","type":"integer","fk":"products"},{"name":"hex","type":"varchar","length":255},{"name":"bin","type":"blob"},{"name":"ip_address","type":"varchar","length":15,"nullable":true}]},{"name":"categories","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"name","type":"varchar","length":255},{"name":"icon","type":"blob","nullable":true}]},{"name":"comments","type":"table","columns":[{"name":"id","type":"bigint","pk":true},{"name":"post_id","type":"integer","fk":"posts"},{"name":"message","type":"varchar","length":255},{"name":"category_id","type":"integer","fk":"categories"}]},{"name":"countries","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"name","type":"varchar","length":255},{"name":"shape","type":"geometry"}]},{"name":"events","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"name","type":"varchar","length":255},{"name":"datetime","type":"timestamp","nullable":true},{"name":"visitors","type":"bigint","nullable":true}]},{"name":"kunsthåndværk","type":"table","columns":[{"name":"id","type":"varchar","length":36,"pk":true},{"name":"Umlauts ä_ö_ü-COUNT","type":"integer"},{"name":"user_id","type":"integer","fk":"users"},{"name":"invisible_id","type":"varchar","length":36,"nullable":true,"fk":"invisibles"}]},{"name":"nopk","type":"table","columns":[{"name":"id","type":"varchar","length":36}]},{"name":"post_tags","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"post_id","type":"integer","fk":"posts"},{"name":"tag_id","type":"integer","fk":"tags"}]},{"name":"posts","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"user_id","type":"integer","fk":"users"},{"name":"category_id","type":"integer","fk":"categories"},{"name":"content","type":"varchar","length":255}]},{"name":"products","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"name","type":"varchar","length":255},{"name":"price","type":"decimal","precision":10,"scale":2},{"name":"properties","type":"clob"},{"name":"created_at","type":"timestamp"},{"name":"deleted_at","type":"timestamp","nullable":true}]},{"name":"tag_usage","type":"view","columns":[{"name":"id","type":"integer","pk":true},{"name":"name","type":"varchar","length":255},{"name":"count","type":"bigint"}]},{"name":"tags","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"name","type":"varchar","length":255},{"name":"is_important","type":"boolean"}]},{"name":"users","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"username","type":"varchar","length":255},{"name":"password","type":"varchar","length":255},{"name":"api_key","type":"varchar","length":255,"nullable":true},{"name":"location","type":"geometry","nullable":true}]}]}`,
			StatusCode: http.StatusOK,
			Driver:     config.Api.Driver,
			SkipFor:    map[string]bool{"sqlite": true},
		},
		{
			Name:       "002_get_barcodes_table",
			Method:     http.MethodGet,
			Uri:        "/columns/barcodes",
			Body:       ``,
			WantJson:   `{"name":"barcodes","type":"table","columns":[{"name":"id","type":"integer","pk":true},{"name":"product_id","type":"integer","fk":"products"},{"name":"hex","type":"varchar","length":255},{"name":"bin","type":"blob"},{"name":"ip_address","type":"varchar","length":15,"nullable":true}]}`,
			StatusCode: http.StatusOK,
		},
		{
			Name:       "003_get_barcodes_id_column",
			Method:     http.MethodGet,
			Uri:        "/columns/barcodes/id",
			Body:       ``,
			WantJson:   `{"name":"id","type":"integer","pk":true}`,
			StatusCode: http.StatusOK,
		},
	}
	utils.RunTests(t, serverUrlHttps, tt)

}
