package cmd

import (
	"errors"
	"fmt"

	"github.com/mikeocool/bbox/core"
	"github.com/spf13/cobra"
)

var SliceCmd = &cobra.Command{
	Use:   "slice",
	Short: "Subdivide the provided box into the nubmer of columns and rows.",
	Args:  cobra.ArbitraryArgs,
	RunE:  runSlice,
}

func runSlice(cmd *cobra.Command, args []string) error {
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

	columns, _ := cmd.Flags().GetInt("columns")
	rows, _ := cmd.Flags().GetInt("rows")
	boxes := bbox.Slice(columns, rows)
	formatted, err := core.FormatCollection(boxes, outputFormat)
	if err != nil {
		return fmt.Errorf("Error formatting result: %v", err)
	}

	// Output the formatted bounding box
	fmt.Println(formatted)
	return nil
}

func init() {
	SliceCmd.Flags().Int("columns", 0, "Number of columns to divide the box into")
	SliceCmd.Flags().Int("rows", 0, "Number of rows to divide the box into")
	SliceCmd.MarkFlagRequired("columns")
	SliceCmd.MarkFlagRequired("rows")
	RootCmd.AddCommand(SliceCmd)
}
