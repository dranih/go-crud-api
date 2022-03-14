package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/record"
)

type FirewallMiddleware struct {
	GenericMiddleware
}

func NewFirewallMiddleware(responder controller.Responder, properties map[string]interface{}) *FirewallMiddleware {
	return &FirewallMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}}
}

func (fwm *FirewallMiddleware) ipMatch(ip, cidr string) bool {
	_, cidrB, err := net.ParseCIDR(cidr)
	if err == nil && cidrB != nil {
		ipA := net.ParseIP(ip)
		if ipA != nil {
			return cidrB.Contains(ipA)
		}
	} else {
		ipA := net.ParseIP(ip)
		ipB := net.ParseIP(cidr)
		if ipA != nil && ipB != nil {
			return ipA.Equal(ipB)
		}
	}
	return false
}

func (fwm *FirewallMiddleware) isIpAllowed(ipAddress, allowedIpAddresses string) bool {
	for _, allowedIp := range strings.Split(allowedIpAddresses, ",") {
		if fwm.ipMatch(ipAddress, allowedIp) {
			return true
		}
	}
	return false
}

func (fwm *FirewallMiddleware) getIpAddress(r *http.Request) string {
	var ipAddress string
	reverseProxy := fmt.Sprint(fwm.getProperty("reverseProxy", ""))
	if reverseProxy != "" {
		ipAddress = r.Header.Get("X-Forwarded-For")
	} else {
		if r.RemoteAddr != "" {
			ipAddress = r.RemoteAddr
		} else {
			ipAddress = "127.0.0.1"
		}
	}
	if ip, _, err := net.SplitHostPort(ipAddress); err == nil {
		return ip
	}
	return ipAddress
}

func (fwm *FirewallMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipAddress := fwm.getIpAddress(r)
		allowedIpAddresses := fwm.getStringProperty("allowedIpAddresses", "")
		if !fwm.isIpAllowed(ipAddress, allowedIpAddresses) {
			log.Printf("Warning - Firewall middleware blocked ip address %s", ipAddress)
			fwm.Responder.Error(record.TEMPORARY_OR_PERMANENTLY_BLOCKED, "", w, "")
			return
		}
		next.ServeHTTP(w, r)
	})
}
