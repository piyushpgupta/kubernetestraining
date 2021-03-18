package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
)

const version = "1.0.0"

var healthCheckResponse = []byte(`{"status":"ready"}`)

// healthCheckHandler is for application heartbeat
func healthCheckHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(healthCheckResponse)
}

// getRemoteIPHandler gets the client/ remote IP address and port
func getRemoteIPHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ip, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Fprintf(w, "user IP address %q is not in the format IP:Port", r.RemoteAddr)
	}
	userIP := net.ParseIP(ip)
	if userIP == nil {
		fmt.Fprintf(w, "user IP address %q is not in the format IP:Port", r.RemoteAddr)
		return
	}

	// Originating IP address of client is coming via a proxy
	forwardIP := r.Header.Get("X-Forwarded-For")

	fmt.Fprintf(w, "IP: %s\n", ip)
	fmt.Fprintf(w, "Port: %s\n", port)
	fmt.Fprintf(w, "Forwarded for IP: %s\n", forwardIP)
}

// helloHandler responds with a "Hello World" message
func helloHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Printf("Serving requests: %s", r.URL.Path)
	hostname, _ := os.Hostname()
	fmt.Fprintf(w, "Hello World!\n")
	fmt.Fprintf(w, "Hostname is: %s\n", hostname)
	fmt.Fprintf(w, "Version is: %s\n", version)
}

func main() {

	// create a httprouter
	router := httprouter.New()

	router.GET("/", helloHandler)
	router.GET("/getip", getRemoteIPHandler)
	router.GET("/health", healthCheckHandler)

	// create a HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// create a channel for signal interrupted
	errChan := make(chan os.Signal, 1)
	signal.Notify(errChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// run the HTTP server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Print("Server is listening on http://localhost:8080/")

	<-errChan
	log.Print("Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer func() {
		// release the resources
		cancel()
	}()

	// Shutdown() gracefully shuts down the server without interrupting any active connections.
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %+v", err)
	}

	log.Print("Server gracefully stopped")
}
