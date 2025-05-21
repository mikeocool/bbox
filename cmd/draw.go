package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"bbox/core"
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
	// Start the drawing server from the core package
	bboxData, err := core.StartDrawServer()
	if err != nil {
		log.Fatalf("Error running draw server: %v", err)
	}

	// If we received data (not interrupted), we can use it here
	if bboxData != nil {
		// Note: The core package already prints the data, but we could do
		// additional processing here if needed
		fmt.Println("Bounding box data received successfully")
	}
}