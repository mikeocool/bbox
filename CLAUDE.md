# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Build and Run
```bash
go build                    # Build the binary
./bbox --draw              # Run with interactive web UI
```

### Testing
```bash
go test -v ./...                    # Run all unit tests
go test -v ./core                   # Run tests for specific package
./integration_tests/setup.sh       # Setup integration test dependencies (run once)
bats integration_tests              # Run integration tests
```

## Architecture Overview

This is a Go CLI tool for bounding box operations with a modular, builder-pattern architecture:

### Core Components

**core/bbox.go**: Central `Bbox` struct with geometric operations (validation, union, buffer, slice, center calculation). All coordinates use Left/Bottom/Right/Top convention.

**input/**: Builder pattern for creating bounding boxes from various sources:
- `RawBuilder`: Parses raw coordinate strings, WKT, GeoJSON
- `FileBuilder`: Loads from shapefiles, GeoJSON files
- `PlaceBuilder`: Geocodes place names using Photon/Komoot service
- `CenterBuilder`: Creates boxes from center point + dimensions
- `BoundsBuilder`: Creates from coordinate bounds with width/height support

**output/**: Formats bounding boxes into various outputs (WKT, GeoJSON, comma-separated, custom Go templates, URLs)

**cmd/**: Cobra CLI structure with subcommands (`center`, `slice`) and root command handling all input/output coordination

**geocoding/**: Place name geocoding integration

### Data Flow
1. Input parsing through builder pattern in `input/input.go`
2. Validation and bbox creation via appropriate builder
3. Optional buffering/drawing operations
4. Output formatting through `output/` package

### Key Design Patterns
- Builder pattern for input handling with validation
- Coordinate validation happens at bbox creation
- Web UI (`--draw`) launches local server for interactive editing
- Pipeline support via stdin/stdout

## Development Notes

The codebase uses standard Go project structure with comprehensive test coverage. Integration tests use BATS framework and require setup script for dependencies.

Buffer operations can expand or shrink bounding boxes. Negative buffer values require validation to prevent invalid geometries.

Drawing interface serves HTML/JS at core/ui/draw.html with real-time coordinate editing.