package cmd

import (
	"errors"
	"fmt"

	"github.com/mikeocool/bbox/core"
	"github.com/spf13/cobra"
)

var CenterCmd = &cobra.Command{
	Use:   "center",
	Short: "Get the center of the bounding box",
	Args:  cobra.ArbitraryArgs,
	RunE:  runCenter,
}

func runCenter(cmd *cobra.Command, args []string) error {
	bbox, err := getBboxFromInput(args)
	if err != nil {
		if errors.Is(err, ErrInputCouldNotCreateBbox) {
			cmd.Usage()
			return err
			// TODO non-zero exit status
		} else {
			return err
		}
	}

	center := bbox.Center()
	formatted, err := core.FormatPoint(center, outputFormat)
	if err != nil {
		return fmt.Errorf("Error formatting point: %v", err)
	}

	// Output the formatted bounding box
	fmt.Println(formatted)
	return nil
}

func init() {
	RootCmd.AddCommand(CenterCmd)
}
