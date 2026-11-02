package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	/*
		The atomic.Int32 type is a really cool standard-library type that allows us to safely increment
		and read an integer value across multiple goroutines (HTTP requests)
	*/
	fileServerHits atomic.Int32
}

// Root Handler
func handlerReadiness(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK) // 200
	writer.Write([]byte(http.StatusText(http.StatusOK)))
}

// apiConfig Struct Middleware method to increment fileServerHits
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		log.Println("Adding one to file server hits..")
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(writer, req)
	})
}

// apiConfig Struct Middleware method to display fileServerHits in response body
func (cfg *apiConfig) middlewareMetricsDis(writer http.ResponseWriter, req *http.Request) {
	log.Println("Reading metrics..")
	writer.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileServerHits.Load())))
}

// apiConfig Struct Middleware method to reset fileServerHits
func (cfg *apiConfig) middlewareMetricsRes(writer http.ResponseWriter, req *http.Request) {
	log.Println("Resetting metrics..")
	writer.Write([]byte("Hits reset to zero"))
	cfg.fileServerHits.Store(0)
}

func main() {
	const filePathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()

	// App (File Server) Handler
	var apiCfg apiConfig
	// Increments metrics by one each app endpoint hit
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	// Metrics Display Handler
	mux.HandleFunc("GET /api/metrics", apiCfg.middlewareMetricsDis) // Display metrics

	// Metrics Reset Handler
	mux.HandleFunc("POST /api/reset", apiCfg.middlewareMetricsRes) // Resets metrics

	// Server def
	server := http.Server{
		Addr:    ":" + port,
		Handler: mux}

	log.Printf("Serving files from %s on port %s\n", filePathRoot, port)
	log.Fatal(server.ListenAndServe())

}
