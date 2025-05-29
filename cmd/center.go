package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/mikeocool/bbox/core"
	"github.com/spf13/cobra"
)

var CenterCmd = &cobra.Command{
	Use:   "center",
	Short: "Get the center of the bounding box",
	Args:  cobra.ArbitraryArgs,
	Run:   runCenter,
}

func runCenter(cmd *cobra.Command, args []string) {
	bbox, err := getBboxFromInput(args)
	if err != nil {
		if errors.Is(err, ErrInputCouldNotCreateBbox) {
			cmd.Usage()
			return
			// TODO non-zero exit status
		} else {
			log.Fatalf("%v", err)
		}
	}

	center := bbox.Center()
	formatted, err := core.FormatPoint(center, outputFormat)
	if err != nil {
		log.Fatalf("Error formatting point: %v", err)
	}

	// Output the formatted bounding box
	fmt.Println(formatted)
}

func init() {
	RootCmd.AddCommand(CenterCmd)
}
