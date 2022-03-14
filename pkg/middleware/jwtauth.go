package middleware

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"hash"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/dranih/go-crud-api/pkg/utils"
)

type JwtAuthMiddleware struct {
	GenericMiddleware
}

func NewJwtAuth(responder controller.Responder, properties map[string]interface{}) *JwtAuthMiddleware {
	return &JwtAuthMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}}
}

func (ja *JwtAuthMiddleware) getAuthorizationToken(r *http.Request) string {
	headerName := ja.getStringProperty("header", "X-Authorization")
	headerValue := r.Header.Get(headerName)
	parts := strings.SplitN(strings.TrimSpace(headerValue), " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}

func (ja *JwtAuthMiddleware) getClaims(token string) map[string]interface{} {
	time := ja.getInt64Property("time", time.Now().Unix())
	leeway := ja.getIntProperty("leeway", 5)
	ttl := ja.getIntProperty("ttl", 30)
	secretsMap := ja.getMapProperty("secrets", "")
	secrets := map[string]string{}
	if len(secretsMap) < 1 {
		secrets = map[string]string{"0": ja.getStringProperty("secret", "")}
	} else {
		i := 0
		for key := range secretsMap {
			secrets[fmt.Sprint(i)] = key
			i++
		}
	}
	requirements := map[string]map[string]bool{
		"alg": ja.getArrayProperty("algorithms", ""),
		"aud": ja.getArrayProperty("audiences", ""),
		"iss": ja.getArrayProperty("issuers", ""),
	}
	return ja.getVerifiedClaims(token, time, leeway, ttl, secrets, requirements)
}

func (ja *JwtAuthMiddleware) getVerifiedClaims(token string, time int64, leeway, ttl int, secrets map[string]string, requirements map[string]map[string]bool) map[string]interface{} {
	algorithms := map[string]string{
		"HS256": "sha256",
		"HS384": "sha384",
		"HS512": "sha512",
		"RS256": "sha256",
		"RS384": "sha384",
		"RS512": "sha512",
	}

	tokenSlice := strings.Split(token, ".")
	if len(token) < 3 {
		return nil
	}

	headerb64 := strings.ReplaceAll(strings.ReplaceAll(tokenSlice[0], `_`, `/`), `-`, `+`)
	headerjson, err := base64.StdEncoding.DecodeString(headerb64)
	if err != nil {
		return nil
	}
	var header map[string]interface{}
	err = json.Unmarshal(headerjson, &header)
	if err != nil {
		return nil
	}
	kid := "0"
	if v, exists := header["kid"]; exists {
		kid = fmt.Sprint(v)
	}
	secret, exists := secrets[kid]
	if !exists {
		return nil
	}
	if v, exists := header["typ"]; !exists || fmt.Sprint(v) != "JWT" {
		return nil
	}

	algorithmI, exists := header["alg"]
	if !exists {
		return nil
	}
	algorithm := fmt.Sprint(algorithmI)
	if requirements["alg"] != nil && len(requirements["alg"]) > 0 && !requirements["alg"][algorithm] {
		return nil
	}

	hmac := algorithms[algorithm]

	signatureb64 := strings.ReplaceAll(strings.ReplaceAll(tokenSlice[2], `_`, `/`), `-`, `+`)
	signature, err := base64.RawStdEncoding.DecodeString(signatureb64)
	if err != nil {
		return nil
	}

	data := fmt.Sprintf("%s.%s", tokenSlice[0], tokenSlice[1])

	switch algorithm[0:1] {
	case "H":
		hash := ja.genHMAC([]byte(data), []byte(secret), hmac)
		if subtle.ConstantTimeCompare(hash, signature) != 1 {
			return nil
		}
	case "R":
		if ja.rsaVerify([]byte(data), signature, []byte(secret), hmac) != 1 {
			return nil
		}
	}
	claimsb64 := strings.ReplaceAll(strings.ReplaceAll(tokenSlice[1], `_`, `/`), `-`, `+`)
	claimsjson, err := base64.StdEncoding.DecodeString(claimsb64)
	if err != nil || claimsjson == nil {
		return nil
	}
	var claims map[string]interface{}
	err = json.Unmarshal(claimsjson, &claims)
	if err != nil {
		return nil
	}

	for field, values := range requirements {
		if values != nil && len(values) > 0 {
			if field == "alg" {
				cfield, exists := claims[field]
				if !exists {
					return nil
				}
				switch t := cfield.(type) {
				case []string:
					found := false
					for _, cf := range t {
						if values[cf] {
							found = true
						}
					}
					if !found {
						return nil
					}
				case string:
					if !values[t] {
						return nil
					}
				}
			}
		}
	}
	nbf, existsNbf := claims["nbf"]
	if existsNbf {
		if a, err := strconv.ParseInt(fmt.Sprint(nbf), 10, 64); err == nil {
			if time+int64(leeway) < a {
				return nil
			}
		}
	}
	iat, existsIat := claims["iat"]
	if existsIat {
		if a, err := strconv.ParseInt(fmt.Sprint(iat), 10, 64); err == nil {
			if time+int64(leeway) < a {
				return nil
			}
		}
	}
	exp, existsExp := claims["exp"]
	if existsExp {
		if a, err := strconv.ParseInt(fmt.Sprint(exp), 10, 64); err == nil {
			if time+int64(leeway) < a {
				return nil
			}
		}
	}
	if existsIat && !existsExp {
		if a, err := strconv.ParseInt(fmt.Sprint(iat), 10, 64); err == nil {
			if time-int64(leeway) > a+int64(ttl) {
				return nil
			}
		}
	}
	return claims
}

func (ja *JwtAuthMiddleware) rsaVerify(data, signature, secret []byte, hmacS string) int {
	var hashl crypto.Hash
	switch hmacS {
	case "sha256":
		hashl = crypto.SHA256
	case "sha384":
		hashl = crypto.SHA384
	case "sha512":
		hashl = crypto.SHA512
	default:
		return -1
	}
	block, _ := pem.Decode([]byte(secret))
	var cert *x509.Certificate
	cert, _ = x509.ParseCertificate(block.Bytes)
	rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)
	err := rsa.VerifyPKCS1v15(rsaPublicKey, hashl, data, signature)
	if err != nil {
		return -1
	}
	return 1
}

func (ja *JwtAuthMiddleware) genHMAC(ciphertext, key []byte, hmacS string) []byte {
	var mac hash.Hash
	switch hmacS {
	case "sha256":
		mac = hmac.New(sha256.New, key)
	case "sha384":
		mac = hmac.New(sha512.New384, key)
	case "sha512":
		mac = hmac.New(sha512.New, key)
	default:
		return nil
	}
	mac.Write([]byte(ciphertext))
	hmac := mac.Sum(nil)
	return hmac
}

func (ja *JwtAuthMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ja.getAuthorizationToken(r)
		session := utils.GetSession(w, r)
		if token != "" {
			claims := ja.getClaims(token)
			if claims == nil || len(claims) < 1 {
				delete(session.Values, "claims")
				session.Save(r, w)
				ja.Responder.Error(record.AUTHENTICATION_FAILED, "JWT", w, "")
				return
			}
			session.Values["claims"] = claims
		}
		if v, exists := session.Values["claims"]; !exists || v == nil {
			if authenticationMode := ja.getProperty("mode", "required"); authenticationMode == "required" {
				realm := fmt.Sprint(ja.getProperty("realm", "Api key required"))
				w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
				ja.Responder.Error(record.AUTHENTICATION_REQUIRED, "", w, realm)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
