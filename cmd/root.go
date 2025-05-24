package cmd

import (
	"fmt"
	"os"

	"bbox/core"

	"github.com/spf13/cobra"
)

// Flag variables
var inputParams core.InputParams
var drawFlag bool
var outputFormat string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "bbox",
	Short: "A CLI application for bounding box operations",
	Long:  `A CLI application that provides tools for working with bounding boxes, including a web-based drawing interface.`,
	Run:   runBase,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	var minX, minY, maxX, maxY float64
	RootCmd.Flags().Float64Var(&minX, "min-x", 0, "Minimum X coordinate of bounding box")
	RootCmd.Flags().Float64Var(&minY, "min-y", 0, "Minimum Y coordinate of bounding box")
	RootCmd.Flags().Float64Var(&maxX, "max-x", 0, "Maximum X coordinate of bounding box")
	RootCmd.Flags().Float64Var(&maxY, "max-y", 0, "Maximum Y coordinate of bounding box")
	RootCmd.Flags().Float64SliceVar(&inputParams.Center, "center", []float64{}, "Center coordinates [x,y] of bounding box")
	RootCmd.Flags().StringVar(&inputParams.Width, "width", "", "Width of bounding box")
	RootCmd.Flags().StringVar(&inputParams.Height, "height", "", "Height of bounding box")
	RootCmd.Flags().StringVar(&inputParams.Raw, "raw", "", "Raw data for bounding box")
	RootCmd.Flags().StringVar(&inputParams.Place, "place", "", "Place name for bounding box")
	RootCmd.Flags().BoolVar(&drawFlag, "draw", false, "Start the drawing interface to create a bounding box")
	RootCmd.Flags().StringVar(&outputFormat, "output", "space", "Output format or destination")

	RootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		// Check if flags were specified and set the pointers accordingly
		if cmd.Flags().Changed("min-x") {
			inputParams.MinX = &minX
		}
		if cmd.Flags().Changed("min-y") {
			inputParams.MinY = &minY
		}
		if cmd.Flags().Changed("max-x") {
			inputParams.MaxX = &maxX
		}
		if cmd.Flags().Changed("max-y") {
			inputParams.MaxY = &maxY
		}
	}
}
