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
	Args:  cobra.ArbitraryArgs,
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
	var left, bottom, right, top float64
	RootCmd.Flags().Float64VarP(&left, "left", "l", 0, "Left coordinate of bounding box")
	RootCmd.Flags().Float64VarP(&bottom, "bottom", "b", 0, "Bottom coordinate of bounding box")
	RootCmd.Flags().Float64VarP(&right, "right", "r", 0, "Right coordinate of bounding box")
	RootCmd.Flags().Float64VarP(&top, "top", "t", 0, "Top coordinate of bounding box")
	RootCmd.Flags().Float64SliceVar(&inputParams.Center, "center", []float64{}, "Center coordinates [x,y] of bounding box")
	RootCmd.Flags().StringVar(&inputParams.Width, "width", "", "Width of bounding box")
	RootCmd.Flags().StringVar(&inputParams.Height, "height", "", "Height of bounding box")
	RootCmd.Flags().StringVar(&inputParams.Place, "place", "", "Place name for bounding box")
	RootCmd.Flags().BoolVar(&drawFlag, "draw", false, "Start the drawing interface to create a bounding box")
	RootCmd.Flags().StringVarP(&outputFormat, "output", "o", "space", "Output format or destination")

	RootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		// Check if flags were specified and set the pointers accordingly
		if cmd.Flags().Changed("left") {
			inputParams.Left = &left
		}
		if cmd.Flags().Changed("bottom") {
			inputParams.Bottom = &bottom
		}
		if cmd.Flags().Changed("right") {
			inputParams.Right = &right
		}
		if cmd.Flags().Changed("top") {
			inputParams.Top = &top
		}
	}
}
