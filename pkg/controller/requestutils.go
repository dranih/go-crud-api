package controller

import (
	"net/http"
)

func getRequestParams(request *http.Request) map[string][]string {
	//params := map[string]string{}
	/*query := request.URL.RawQuery
	query = strings.Replace(strings.Replace(query, `=`, `%5B%5D=`, -1), `%5D%5B%5D=`, `%5D=`, -1)
	params, _ := url.ParseQuery(query)*/
	return request.URL.Query()
}

/*
			$query = $request->getUri()->getQuery();
            //$query = str_replace('][]=', ']=', str_replace('=', '[]=', $query));
            $query = str_replace('%5D%5B%5D=', '%5D=', str_replace('=', '%5B%5D=', $query));
            parse_str($query, $params);
            return $params;
*/
