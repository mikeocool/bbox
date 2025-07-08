package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/mikeocool/bbox/core"
	"github.com/mikeocool/bbox/input"
	"github.com/mikeocool/bbox/logger"
	"github.com/mikeocool/bbox/output"

	"github.com/spf13/cobra"
)

// Flag variables
var inputParams input.InputParams
var drawFlag bool
var verboseFlag bool
var addressFlag string
var portFlag int
var outputSettings output.OutputSettings

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

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Enable debug logging if verbose flag is set
		if verboseFlag {
			logger.EnableDebug()
		}

		// If the flags have had values passed in set them
		inputParams.Left = GetFlagFloat64(cmd, "left")
		inputParams.Bottom = GetFlagFloat64(cmd, "bottom")
		inputParams.Top = GetFlagFloat64(cmd, "top")
		inputParams.Right = GetFlagFloat64(cmd, "right")

		outputFlag := cmd.Flag("output")
		format, details := output.ParseFormat(outputFlag.Value.String())
		outputSettings.FormatType = format
		outputSettings.FormatDetails = details
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		if userErr, ok := err.(*UserError); ok {
			fmt.Fprintf(os.Stderr, "Error: %v\n", userErr.UserError())
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

func GetFlagFloat64(cmd *cobra.Command, flagName string) *float64 {
	flag := cmd.Flag(flagName)
	if flag != nil && flag.Changed {
		val, _ := cmd.Flags().GetFloat64(flagName)
		return &val
	}
	return nil
}

func init() {
	// input flags
	RootCmd.PersistentFlags().Float64P("left", "l", 0, "Left coordinate of bounding box")
	RootCmd.PersistentFlags().Float64P("bottom", "b", 0, "Bottom coordinate of bounding box")
	RootCmd.PersistentFlags().Float64P("right", "r", 0, "Right coordinate of bounding box")
	RootCmd.PersistentFlags().Float64P("top", "t", 0, "Top coordinate of bounding box")
	RootCmd.PersistentFlags().Float64SliceVar(&inputParams.Center, "center", []float64{}, "Center coordinates [x,y] of bounding box")
	RootCmd.PersistentFlags().StringVar(&inputParams.Width, "width", "", "Width of bounding box")
	RootCmd.PersistentFlags().StringVar(&inputParams.Height, "height", "", "Height of bounding box")
	RootCmd.PersistentFlags().StringVar(&inputParams.Place, "place", "", "Place name for bounding box")
	RootCmd.PersistentFlags().StringVar(&inputParams.Geocoder, "geocoder", "", "Geocoder service to use (requires --place)")
	RootCmd.PersistentFlags().StringVar(&inputParams.GeocoderURL, "geocoder-url", "", "Custom geocoder URL with %s placeholder for place name (requires --place)")
	RootCmd.PersistentFlags().StringSliceVar(&inputParams.GeocoderHeaders, "geocoder-header", []string{}, "HTTP headers for geocoder requests in 'Name: Value' format (can be used multiple times)")
	//RootCmd.PersistentFlags().StringSliceVarP(&inputParams.File, "file", "f", []string{}, "Path to file to load")

	RootCmd.PersistentFlags().Float64Var(&inputParams.Buffer, "buffer", 0, "Grow the box by the specified amount, or shrink it if the value is negative.")
	RootCmd.PersistentFlags().IntVar(&inputParams.SridOverride, "srid", 0, "Override the SRID of the input bounding box")

	RootCmd.PersistentFlags().BoolVar(&drawFlag, "draw", false, "Start the drawing interface to create a bounding box")
	RootCmd.PersistentFlags().StringVar(&addressFlag, "address", "localhost", "IP address to bind the draw server to")
	RootCmd.PersistentFlags().IntVar(&portFlag, "port", 0, "Port to bind the draw server to (0 = auto-find available port)")
	RootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose debug logging")

	RootCmd.PersistentFlags().StringP("output", "o", "space", "Output format or destination")
	RootCmd.PersistentFlags().IntVar(&outputSettings.GeojsonIndent, "geojson-indent", 0, "Indentation level for geojson output format")
	RootCmd.PersistentFlags().StringVar(&outputSettings.GeojsonType, "geojson-type", "", "Type of geojson object to output - featurecollection, feature, geometry, or coordinates")
	RootCmd.PersistentFlags().BoolVar(&outputSettings.Browser, "browser", false, "Open URL in browser (only valid with URL output format)")
}

var ErrInputCouldNotCreateBbox = errors.New("could not create bounding box")

func getBboxFromInput(args []string) (core.Bbox, error) {
	// Create a bounding box from input parameters
	if input.IsInputFromPipe() {
		inputParams.DataStream = os.Stdin
	} else if len(args) > 0 {
		inputParams.RawArgs = args
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
			return core.Bbox{}, fmt.Errorf("error creating bounding box: %w", err)
		}
	}

	if drawFlag {
		// Start the drawing server
		bbox, err = core.StartDrawServer(bbox, addressFlag, portFlag)
		if err != nil {
			if errors.Is(err, core.ErrNonWGS84Coordinates) {
				return core.Bbox{}, NewUserError(err, "Box coordinates appear to be outside of the range of valid WGS84 coordinates. Cannot show non-WGS84 coordinates in --draw mode")
			} else if errors.Is(err, core.ErrUnsupportedSRID) {
				return core.Bbox{}, NewUserError(err, "Draw mode does not support the specified SRID. Only geographic coordinate systems (WGS84/4326, NAD83/4269) are supported in --draw mode")
			}
			return core.Bbox{}, fmt.Errorf("starting draw server: %w", err)
		}
	}

	return bbox, nil
}

func handleOutput(bbox core.Bbox, settings output.OutputSettings) error {
	formatted, err := output.FormatBbox(bbox, outputSettings)
	if err != nil {
		return fmt.Errorf("error formatting bounding box: %w", err)
	}

	fmt.Println(formatted)

	// Open browser if requested
	// TODO suppport template as well?
	if settings.FormatType == output.FormatUrl && settings.Browser {
		core.OpenBrowser(formatted)
	}

	return nil
}

func runRoot(cmd *cobra.Command, args []string) error {
	// Validate output settings
	if err := outputSettings.Validate(); err != nil {
		return err
	}

	bbox, err := getBboxFromInput(args)
	if err != nil {
		if errors.Is(err, ErrInputCouldNotCreateBbox) {
			cmd.Usage()
			return err
		} else {
			return err
		}
	}

	if err := handleOutput(bbox, outputSettings); err != nil {
		return fmt.Errorf("error formatting bounding box: %w", err)
	}

	return nil
}
