package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

// drawCmd represents the draw command
var DrawCmd = &cobra.Command{
	Use:   "draw",
	Short: "Start a web server to draw a bounding box",
	Long:  `Start a web server that serves a UI for drawing a bounding box on a map. The server will listen on the first available port starting from 5000.`,
	Run:   runDraw,
}

func init() {
	// Register this command with the root command
	RootCmd.AddCommand(DrawCmd)
}

func runDraw(cmd *cobra.Command, args []string) {
	// Find the first available port starting from 5000
	port := findAvailablePort(5000)
	if port == 0 {
		log.Fatal("Could not find an available port")
	}

	// Create a server with the UI handler
	mux := http.NewServeMux()

	// Serve static files from the ui/draw directory
	uiDir := filepath.Join("ui", "draw")
	fileServer := http.FileServer(http.Dir(uiDir))
	mux.Handle("/", fileServer)

	// Create a channel to receive the bounding box data
	bboxCh := make(chan []byte)

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

		// Send the body to the channel
		bboxCh <- body

		// Return a success response
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Create a channel for context cancelation
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the server in a goroutine
	go func() {
		fmt.Printf("Go to http://localhost:%d to draw your bounding box\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Open the browser
	go func() {
		time.Sleep(500 * time.Millisecond) // Give the server a moment to start
		openBrowser(fmt.Sprintf("http://localhost:%d", port))
	}()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	// Wait for either the bbox data or a signal
	select {
	case body := <-bboxCh:
		// Print the received data
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
			fmt.Println("Received bounding box data:")
			fmt.Println(string(body))
		} else {
			fmt.Println("Received bounding box data:")
			fmt.Println(prettyJSON.String())
		}
	case <-sigCh:
		fmt.Println("Interrupted by user")
	}

	// Shutdown the server
	fmt.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
	fmt.Println("Server stopped")
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