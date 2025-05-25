package core

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"time"
)

//go:embed ui/draw.html
var drawHTML []byte

// DrawServer holds the configuration and state for the bounding box drawing server
type DrawServer struct {
	Port int
}

type TemplateContext struct {
	Bbox Bbox
}

// StartDrawServer starts a web server for drawing bounding boxes.
// It returns the received bounding box data as a Bbox struct.
func StartDrawServer(bbox Bbox) (Bbox, error) {
	// Find the first available port starting from 5000
	port := findAvailablePort(5000)
	if port == 0 {
		return Bbox{}, fmt.Errorf("could not find an available port")
	}

	server := &DrawServer{
		Port: port,
	}

	return server.Start(bbox)
}

// Start starts the web server and returns the bounding box data when received
func (s *DrawServer) Start(inputBbox Bbox) (Bbox, error) {
	// Create a server with the UI handler
	mux := http.NewServeMux()

	// Parse the HTML template
	tmpl, err := template.New("draw").Parse(string(drawHTML))
	if err != nil {
		return Bbox{}, fmt.Errorf("failed to parse template: %v", err)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		err := tmpl.Execute(w, TemplateContext{
			Bbox: inputBbox,
		})
		if err != nil {
			log.Printf("Failed to execute template: %v", err)
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			return
		}
	})

	// Create a channel to receive the bounding box data
	bboxCh := make(chan Bbox)

	// Add the /bbox endpoint
	mux.HandleFunc("/bbox", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		// Parse the JSON into a Bbox struct
		var bbox Bbox
		if err := json.Unmarshal(body, &bbox); err != nil {
			http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Validate that the bbox is valid before proceeding
		if err := bbox.Validate(); err != nil {
			http.Error(w, "Invalid bounding box: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Send the bbox struct directly to the channel
		bboxCh <- bbox

		// Return a success response
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: mux,
	}

	// Create a context for graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the server in a goroutine
	go func() {
		fmt.Printf("Go to http://localhost:%d to draw your bounding box\n", s.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Open the browser
	go func() {
		time.Sleep(500 * time.Millisecond) // Give the server a moment to start
		openBrowser(fmt.Sprintf("http://localhost:%d", s.Port))
	}()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	var bbox Bbox

	// Wait for either the bbox data or a signal
	select {
	case bbox = <-bboxCh:
		fmt.Println("Received bounding box data")
	case <-sigCh:
		fmt.Println("Interrupted by user")
	}

	// Shutdown the server
	fmt.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return Bbox{}, fmt.Errorf("server shutdown error: %v", err)
	}
	fmt.Println("Server stopped")

	return bbox, nil
}

// findAvailablePort returns the first available port starting from the given port
func findAvailablePort(startPort int) int {
	for port := startPort; port < startPort+1000; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port
		}
	}
	return 0
}

// openBrowser opens the default browser to the specified URL
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}
