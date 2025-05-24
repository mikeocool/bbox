package cmd

import (
	"errors"
	"fmt"
	"log"

	"bbox/core"

	"github.com/spf13/cobra"
)

func runBase(cmd *cobra.Command, args []string) {
	// Create a bounding box from input parameters
	bbox, err := inputParams.GetBbox()
	if err != nil {
		var noUsableBuilderError core.NoUsableBuilderError
		if errors.As(err, &noUsableBuilderError) {
			// If no usable builder is found and we're not drawing, print usage and exit
			if !drawFlag {
				cmd.Usage()
				return
			}
		} else {
			log.Fatalf("Error creating bounding box: %v", err)
		}
	}

	if drawFlag {
		// Start the drawing server
		// TODO pass in bbox is it's set
		bbox, err = core.StartDrawServer()
		if err != nil {
			log.Fatalf("Error running draw server: %v", err)
		}
	}

	formatted, err := core.Format(bbox, outputFormat)
	if err != nil {
		log.Fatalf("Error formatting bounding box: %v", err)
	}

	// Output the formatted bounding box
	fmt.Println(formatted)
}
