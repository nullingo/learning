package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
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
	fmt.Println("code 200", getRemoteIp((req)))
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

	g, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	g.Go(func() error {
		<-ctx.Done()
		fmt.Println("context done, http shutdown")
		return server.Shutdown(context.TODO())
	})

	g.Go(func() error {
		defer cancel()
		return server.ListenAndServe()
	})

	g.Go(func() error {
		exitSigs := []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT}
		sig := make(chan os.Signal, len(exitSigs))
		signal.Notify(sig, exitSigs...)
		for {
			select {
			case <-ctx.Done():
				fmt.Println("context done")
				return ctx.Err()
			case s := <-sig:
				fmt.Println("capture os signal:", s)
				cancel()
				return nil
			}
		}
	})

	err := g.Wait()
	fmt.Println(err)
}
