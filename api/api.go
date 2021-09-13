package api

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"node/config"
)

// RPC Api structure
type Api struct{}

type HttpConn struct {
	io.Reader
	io.Writer
}

func (c *HttpConn) Close() error { return nil }

// Start server with Test instance as a service
func ServerStart() {
	server := rpc.NewServer()
	_ = server.Register(&Api{})

	listener, err := net.Listen("tcp", ":"+config.ApiPort)

	log.Println("Api server started.")

	if err != nil {
		log.Fatal("listen error:", err)
	}

	defer listener.Close()

	go cleanRequestsMemory()

	_ = http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api" {

			// Ограничение количества запросов к апи
			if requestInRequestsMemory(r.Header, r.Method) {
				http.Error(setHeaders(w, 0), "Too many requests. Try later.", 429)
				return
			}

			if err := server.ServeRequest(jsonrpc.NewServerCodec(&HttpConn{r.Body, setHeaders(w, 0)})); err != nil {
				log.Printf("Error while serving JSON request: %v", err)
				http.Error(w, "Error while serving JSON request, details have been logged.", 500)
				return
			}
		} else {
			http.Error(w, "Unknown request", 404)
		}
	}))
}

func setHeaders(w http.ResponseWriter, code int) http.ResponseWriter {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	if code != 0 {
		w.WriteHeader(code)
	}

	return w
}
