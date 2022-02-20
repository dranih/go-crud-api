package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	mxj "github.com/clbanning/mxj/v2"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("4d70ad3e4165e3a1dc158fdc0fc07dc9ae8c8c4ecbd7f8619debebcba5a3710d"))

func GetRequestParams(request *http.Request) url.Values {
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

//GetBodyData tries to get data from body request, as a urlencoded content type or as json by default
func GetBodyData(r *http.Request) (interface{}, error) {
	headerContentType := r.Header.Get("Content-Type")
	if headerContentType == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		res := map[string]interface{}{}
		for key, val := range r.PostForm {
			if strings.HasSuffix(key, "__is_null") {
				res[strings.TrimSuffix(key, "__is_null")] = nil
			} else {
				res[key] = strings.Join(val, ",")
			}
		}
		return res, nil
	} else {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		r.Body.Close() //  must close
		r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		var jsonMap interface{}

		if headerContentType == "application/xml" {
			var mv mxj.Map
			mv, err = mxj.NewMapXml(b, true)
			if err != nil {
				return nil, err
			}
			root, err := mv.Root()
			if err != nil {
				return nil, err
			}
			jsonMap, err = mv.ValueForKey(root)
			if err != nil {
				return nil, err
			}
			if val, ok := jsonMap.(map[string]interface{}); ok {
				if val1, exists := val["object"]; exists {
					if array, exists := val["-type"]; exists && array == "array" {
						jsonMap = val1
					}
				}
			}
		} else {
			err = json.Unmarshal(b, &jsonMap)
			if err != nil {
				return nil, err
			}
		}

		return jsonMap, nil
	}
}

func GetBodyMapData(r *http.Request) (map[string]interface{}, error) {
	if res, err := GetBodyData(r); err != nil {
		return nil, err
	} else if resMap, ok := res.(map[string]interface{}); !ok {
		return nil, errors.New("unable to decode body")
	} else {
		return resMap, nil
	}
}

func GetPathSegment(r *http.Request, part int) string {
	path := r.URL.Path
	pathSegments := strings.Split(strings.TrimRight(path, "/"), "/")
	if part < 0 || part >= len(pathSegments) {
		return ""
	}
	return pathSegments[part]
}

func GetOperation(r *http.Request) string {
	method := r.Method
	path := GetPathSegment(r, 1)
	hasPk := false
	if GetPathSegment(r, 3) != "" {
		hasPk = true
	}
	switch path {
	case "openapi":
		return "document"
	case "columns":
		if method == "get" {
			return "reflect"
		} else {
			return "remodel"
		}
	case "geojson":
	case "records":
		switch method {
		case "POST":
			return "create"
		case "GET":
			if hasPk {
				return "read"
			} else {
				return "list"
			}
		case "PUT":
			return "update"
		case "DELETE":
			return "delete"
		case "PATCH":
			return "increment"
		}
	}
	return "unknown"
}

func GetTableNames(r *http.Request, allTableNames []string) []string {
	path := GetPathSegment(r, 1)
	tableName := GetPathSegment(r, 2)
	switch path {
	case "openapi":
		return allTableNames
	case "columns":
		if tableName != "" {
			return []string{tableName}
		} else {
			return allTableNames
		}
	case "records":
		return getJoinTables(tableName, GetRequestParams(r))
	}
	return allTableNames
}

func getJoinTables(tableName string, parameters map[string][]string) []string {
	uniqueTableNames := map[string]bool{}
	uniqueTableNames[tableName] = true
	if join, exists := parameters["join"]; exists {
		for _, parameter := range join {
			tableNames := strings.Split(strings.TrimSpace(parameter), ",")
			for _, tableNamef := range tableNames {
				uniqueTableNames[tableNamef] = true
			}
		}
	}
	var keys []string
	for key := range uniqueTableNames {
		keys = append(keys, key)
	}
	return keys
}

//CertSetup generate a self signed certificate if https is on and no certificate is provided
//To be used for development purposes only
//From https://gist.github.com/shaneutt/5e1995295cff6721c89a71d13a71c251 https://shaneutt.com/blog/golang-ca-and-signed-cert-go/
func CertSetup(ipAddress string) (serverTLSConf *tls.Config, clientTLSConf *tls.Config, err error) {
	var ips []net.IP
	if ipAddress != "" {
		if ip := net.ParseIP(ipAddress); ip != nil {
			ips = append(ips, ip)
		}
	}
	if len(ips) < 1 {
		ips = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
	}

	// set up our CA certificate
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2022),
		Subject: pkix.Name{
			Organization:  []string{"Company, INC."},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// create our private and public key
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	// pem encode
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2022),
		Subject: pkix.Name{
			Organization:  []string{"Company, INC."},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
		},
		IPAddresses:  ips,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
	if err != nil {
		return nil, nil, err
	}

	serverTLSConf = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(caPEM.Bytes())
	clientTLSConf = &tls.Config{
		RootCAs: certpool,
	}

	return
}
