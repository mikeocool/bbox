package cmd

import (
	"fmt"
	"log"

	"bbox/core"
	"github.com/spf13/cobra"
)

func runBase(cmd *cobra.Command, args []string) {
	if drawFlag {
		// Start the drawing server
		bbox, err := core.StartDrawServer()
		if err != nil {
			log.Fatalf("Error running draw server: %v", err)
		}

		// Format the bounding box using SpaceFormat
		formatted, err := core.SpaceFormat(bbox)
		if err != nil {
			log.Fatalf("Error formatting bounding box: %v", err)
		}

		// Output the formatted bounding box
		fmt.Println(formatted)
	} else {
		fmt.Println("Root")
	}
}
