package middleware

import (
	"bufio"
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

func (bam *BasicAuthMiddleware) getAuthorizationCredentials(request *http.Request) (string, string, bool) {
	return request.BasicAuth()
}

func (bam *BasicAuthMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := utils.SetSession(w, r)
		username, password, ok := bam.getAuthorizationCredentials(r)
		if ok {
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
