package middleware

import (
	"bufio"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/dranih/go-crud-api/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type BasicAuthMiddleware struct {
	GenericMiddleware
}

func (bam *BasicAuthMiddleware) hasCorrectPassword(username, password string, passwords *map[string]string) (bool, bool) {
	rewrite := false
	if hash, exists := (*passwords)[username]; exists {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		if cost, errCost := bcrypt.Cost([]byte(password)); errCost == nil && cost < bcrypt.MinCost {
			if newHash, errNew := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); errNew == nil {
				(*passwords)[username] = string(newHash)
				rewrite = true
			}
		}
		return err == nil, rewrite
	}
	return false, rewrite

}

func (bam *BasicAuthMiddleware) getValidUsername(username, password, passwordFile string) string {
	passwords, rewrite1 := bam.readPasswords(passwordFile)
	valid, rewrite2 := bam.hasCorrectPassword(username, password, &passwords)
	if rewrite1 || rewrite2 {
		if err := bam.writePasswords(passwordFile, passwords); err != nil {
			log.Printf("Error writing password file %s : %v", passwordFile, err)
		}
	}
	if valid {
		return username
	}
	return ""
}

func (bam *BasicAuthMiddleware) readPasswords(passwordFile string) (map[string]string, bool) {
	rewrite := false
	passwords := map[string]string{}
	file, err := os.Open(passwordFile)
	if err != nil {
		log.Printf("Error opening %s : %v", passwordFile, err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if text := scanner.Text(); strings.Contains(text, ":") {
			val := strings.SplitN(strings.TrimSpace(text), ":", 2)
			if len(val[1]) > 0 && val[1][0:1] != `$` {
				hash, err := bcrypt.GenerateFromPassword([]byte(val[1]), bcrypt.DefaultCost)
				if err == nil {
					val[1] = string(hash)
					rewrite = true
				}
			}
			passwords[val[0]] = val[1]
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading %s : %v", passwordFile, err)
	}
	return passwords, rewrite
}

func (bam *BasicAuthMiddleware) writePasswords(passwordFile string, passwords map[string]string) error {
	file, err := os.Create(passwordFile)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	for username, hash := range passwords {
		_, err := writer.WriteString(username + ":" + hash + "\n")
		if err != nil {
			return err
		}
	}
	writer.Flush()
	return nil
}

func (bam *BasicAuthMiddleware) getAuthorizationCredentials(request *http.Request) string {
	serverParams := utils.GetRequestParams(request)
	if authUser, exists := serverParams["AUTH_USER"]; exists {
		if authPwd, exists := serverParams["AUTH_PW"]; exists {
			return authUser[0] + ":" + authPwd[0]
		}
	}
	header := request.Header.Get("Authorization")
	parts := strings.Split(strings.TrimSpace(header), " ")
	if len(parts) != 2 {
		return ""
	}
	if parts[0] != "Basic" {
		return ""
	}
	if sDec, err := base64.StdEncoding.DecodeString(parts[1]); err != nil {
		return string(sDec)
	}
	return ""
}

func (bam *BasicAuthMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		session := utils.SetSession(w, r)
		credentials := bam.getAuthorizationCredentials(r)
		if credentials != "" {
			var username, password string
			if strings.Contains(credentials, ":") {
				val := strings.SplitN(strings.TrimSpace(credentials), ":", 2)
				username = val[0]
				password = val[1]
			}
			passwordFile := bam.getProperty("passwordFile", ".htpasswd")
			validUser := bam.getValidUsername(username, password, passwordFile)
			session.Values["username"] = validUser
			if validUser == "" {
				bam.Responder.Error(record.AUTHENTICATION_FAILED, username, w, "")
			} else {
				if err := session.Save(r, w); err != nil {
					bam.Responder.Error(record.INTERNAL_SERVER_ERROR, err.Error(), w, "")
				} else {
					next.ServeHTTP(w, r)
				}
			}
		} else {
			if val, exists := session.Values["username"]; !exists || val == nil {
				if authenticationMode := bam.getProperty("mode", "required"); authenticationMode == "required" {
					realm := bam.getProperty("realm", "Username and password required")
					w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
					bam.Responder.Error(record.AUTHENTICATION_REQUIRED, "", w, "")
				} else {
					next.ServeHTTP(w, r)
				}
			}
			next.ServeHTTP(w, r)
		}
	})
}

/*
	public function process(ServerRequestInterface $request, RequestHandlerInterface $next): ResponseInterface
	{
		if (session_status() == PHP_SESSION_NONE) {
			if (!headers_sent()) {
				$sessionName = $this->getProperty('sessionName', '');
				if ($sessionName) {
					session_name($sessionName);
				}
				session_start();
			}
		}
		$credentials = $this->getAuthorizationCredentials($request);
		if ($credentials) {
			list($username, $password) = array('', '');
			if (strpos($credentials, ':') !== false) {
				list($username, $password) = explode(':', $credentials, 2);
			}
			$passwordFile = $this->getProperty('passwordFile', '.htpasswd');
			$validUser = $this->getValidUsername($username, $password, $passwordFile);
			$_SESSION['username'] = $validUser;
			if (!$validUser) {
				return $this->responder->error(ErrorCode::AUTHENTICATION_FAILED, $username);
			}
			if (!headers_sent()) {
				session_regenerate_id();
			}
		}
		if (!isset($_SESSION['username']) || !$_SESSION['username']) {
			$authenticationMode = $this->getProperty('mode', 'required');
			if ($authenticationMode == 'required') {
				$response = $this->responder->error(ErrorCode::AUTHENTICATION_REQUIRED, '');
				$realm = $this->getProperty('realm', 'Username and password required');
				$response = $response->withHeader('WWW-Authenticate', "Basic realm=\"$realm\"");
				return $response;
			}
		}
		return $next->handle($request);
	}
}*/
