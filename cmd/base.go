package cmd

import (
	"fmt"
	"log"

	"bbox/core"

	"github.com/spf13/cobra"
)

func runBase(cmd *cobra.Command, args []string) {
	// Create a bounding box from input parameters
	bbox, err := inputParams.GetBbox()
	if err != nil {
		log.Fatalf("Error creating bounding box: %v", err)
	}

	if drawFlag {
		// Start the drawing server
		bbox, err = core.StartDrawServer()
		if err != nil {
			log.Fatalf("Error running draw server: %v", err)
		}
	}

	formatted, err := core.SpaceFormat(bbox)
	if err != nil {
		log.Fatalf("Error formatting bounding box: %v", err)
	}

	// Output the formatted bounding box
	fmt.Println(formatted)
}
