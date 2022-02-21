package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
)

func getRemoteIp(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get(XRealIP); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get(XForwardedFor); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

type headerHandler struct {
}

func (h *headerHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	for k := range req.Header {
		resp.Header().Add(k, req.Header.Get(k))
	}
	version, ok := os.LookupEnv("VERSION")
	if ok {
		resp.Header().Add("VERSION", version)
	}
	fmt.Println("code 200", getRemoteIp((req)))
}

type healthHandler struct {
}

func (h *healthHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
	fmt.Println("code 200")
}

func main() {
	mux := http.NewServeMux()

	header := &headerHandler{}
	mux.Handle("/", header)

	health := &healthHandler{}
	mux.Handle("/localhost/healthz", health)

	server := &http.Server{
		Addr:    "127.0.0.1:8000",
		Handler: mux,
	}
	server.ListenAndServe()
}
