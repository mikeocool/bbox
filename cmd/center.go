package cmd

import (
	"errors"
	"log"

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
	log.Printf("%f %f", center[0], center[1]) // TODO output as point
}

func init() {
	RootCmd.AddCommand(CenterCmd)
}
