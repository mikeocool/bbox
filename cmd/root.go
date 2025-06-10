package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mikeocool/bbox/core"
	"github.com/mikeocool/bbox/input"

	"github.com/spf13/cobra"
)

// Flag variables
var inputParams input.InputParams
var drawFlag bool
var outputFormat string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "bbox",
	Short: "A CLI application for bounding box operations",
	Long:  `A CLI application that provides tools for working with bounding boxes, including a web-based drawing interface.`,
	Args:  cobra.ArbitraryArgs,

	// we'll manage printing errors and usage orselves
	// cobra does it in a lot of cases where we dont want it
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: runRoot,
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
	RootCmd.PersistentFlags().Float64VarP(&left, "left", "l", 0, "Left coordinate of bounding box")
	RootCmd.PersistentFlags().Float64VarP(&bottom, "bottom", "b", 0, "Bottom coordinate of bounding box")
	RootCmd.PersistentFlags().Float64VarP(&right, "right", "r", 0, "Right coordinate of bounding box")
	RootCmd.PersistentFlags().Float64VarP(&top, "top", "t", 0, "Top coordinate of bounding box")
	RootCmd.PersistentFlags().Float64SliceVar(&inputParams.Center, "center", []float64{}, "Center coordinates [x,y] of bounding box")
	RootCmd.PersistentFlags().StringVar(&inputParams.Width, "width", "", "Width of bounding box")
	RootCmd.PersistentFlags().StringVar(&inputParams.Height, "height", "", "Height of bounding box")
	RootCmd.PersistentFlags().StringVar(&inputParams.Place, "place", "", "Place name for bounding box")
	RootCmd.PersistentFlags().StringSliceVarP(&inputParams.File, "file", "f", []string{}, "Path to file to load")
	RootCmd.PersistentFlags().BoolVar(&drawFlag, "draw", false, "Start the drawing interface to create a bounding box")
	RootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "space", "Output format or destination")

	RootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		// Check if flags were specified and set the pointers accordingly
		if cmd.PersistentFlags().Changed("left") {
			inputParams.Left = &left
		}
		if cmd.PersistentFlags().Changed("bottom") {
			inputParams.Bottom = &bottom
		}
		if cmd.PersistentFlags().Changed("right") {
			inputParams.Right = &right
		}
		if cmd.PersistentFlags().Changed("top") {
			inputParams.Top = &top
		}
	}
}

var ErrInputCouldNotCreateBbox = errors.New("could not create bounding box")

func getBboxFromInput(args []string) (core.Bbox, error) {
	// Create a bounding box from input parameters
	if input.IsInputFromPipe() {
		stdinBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return core.Bbox{}, fmt.Errorf("Error reading from stdin: %w", err)
		}
		inputParams.Raw = stdinBytes
	} else if len(args) > 0 {
		inputParams.Raw = []byte(strings.Join(args, " "))
	}

	bbox, err := inputParams.GetBbox()
	if err != nil {
		var noUsableBuilderError input.NoUsableBuilderError
		if errors.As(err, &noUsableBuilderError) {
			// If no usable builder is found and we're not drawing, print usage and exit
			if !drawFlag {
				return core.Bbox{}, ErrInputCouldNotCreateBbox
			}
		} else {
			return core.Bbox{}, fmt.Errorf("Error creating bounding box: %w", err)
		}
	}

	if drawFlag {
		// Start the drawing server
		bbox, err = core.StartDrawServer(bbox)
		if err != nil {
			return core.Bbox{}, fmt.Errorf("Error running draw server: %w", err)
		}
	}

	return bbox, nil
}

func runRoot(cmd *cobra.Command, args []string) error {
	bbox, err := getBboxFromInput(args)
	if err != nil {
		if errors.Is(err, ErrInputCouldNotCreateBbox) {
			return err
			// TODO non-zero exit status
		} else {
			return err
		}
	}

	formatted, err := core.FormatBbox(bbox, outputFormat)
	if err != nil {
		return fmt.Errorf("Error formatting bounding box: %w", err)
	}

	// Output the formatted bounding box
	fmt.Println(formatted)
	return nil
}
