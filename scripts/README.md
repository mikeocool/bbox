# Scripts Directory

This directory contains standalone utility scripts for the bbox project.

## Scripts

### generate_parquet.go

Generates a Parquet file with sample data including WKB-encoded point geometries. The file contains several string columns (id, name, city, region) and a `geom` column with points encoded in Well-Known Binary (WKB) format.

**Note**: This script generates a standard Parquet file without GeoParquet metadata - it's just raw WKB bytes in a binary column.

#### Dependencies

First, add the parquet-go dependency to your project:

```bash
go get github.com/parquet-go/parquet-go
```

#### Usage

```bash
# Run from project root
go run scripts/generate_parquet.go [flags]
```

#### Flags

- `--output`: Output file path (default: "test_data.parquet")
- `--count`: Number of records to generate (default: 1000)
- `--seed`: Random seed for reproducible data (default: 0, uses current time)

#### Examples

```bash
# Generate 1000 records to default file
go run scripts/generate_parquet.go

# Generate 5000 records to specific file
go run scripts/generate_parquet.go --output data/sample_points.parquet --count 5000

# Generate with specific seed for reproducible data
go run scripts/generate_parquet.go --seed 12345 --count 100
```

#### Output Schema

The generated Parquet file has the following schema:

| Column | Type   | Description |
|--------|--------|-------------|
| id     | string | Unique identifier (ID_000001, ID_000002, etc.) |
| name   | string | Random person name |
| city   | string | Random US city name |
| region | string | Random US region name |
| geom   | binary | WKB-encoded point geometry |

#### WKB Format

The `geom` column contains points encoded in Well-Known Binary format:
- Byte order: Little endian
- Geometry type: Point (1)
- Coordinates: Random points within continental US bounds
  - Longitude: -125.0 to -75.0
  - Latitude: 25.0 to 50.0

#### Reading the Generated File

You can read the generated Parquet file using various tools:

**Using parquet-tools (if installed):**
```bash
parquet-tools head test_data.parquet
```

**Using Python with geopandas:**
```python
import pandas as pd
import geopandas as gpd
from shapely import wkb

# Read parquet file
df = pd.read_parquet('test_data.parquet')

# Convert WKB to geometry
df['geometry'] = df['geom'].apply(lambda x: wkb.loads(bytes(x)))
gdf = gpd.GeoDataFrame(df.drop('geom', axis=1), geometry='geometry')
```

**Using Go:**
```go
// Read back with parquet-go
file, _ := os.Open("test_data.parquet")
reader := parquet.NewGenericReader[Record](file)
records := make([]Record, 100)
n, _ := reader.Read(records)
```

### verify_parquet.go

Verifies and inspects the contents of generated Parquet files, displaying sample records and statistics.

#### Usage

```bash
# Run from project root
go run scripts/verify_parquet.go [flags]
```

#### Flags

- `--input`: Input parquet file path (default: "test_data.parquet")
- `--all`: Show all records instead of just first 10
- `--max`: Maximum records to display when not using --all (default: 10)

#### Examples

```bash
# Verify default file
go run scripts/verify_parquet.go

# Verify specific file
go run scripts/verify_parquet.go --input data/sample_points.parquet

# Show all records
go run scripts/verify_parquet.go --input sample.parquet --all

# Show first 5 records only
go run scripts/verify_parquet.go --max 5
```

#### Output

The verification script provides:
- Total record count
- Sample records with decoded WKB coordinates
- Statistics including coordinate ranges
- Unique cities and regions counts
- Top cities and regions by frequency

### demo_parquet.sh

A shell script that demonstrates the complete workflow of generating and verifying Parquet files. This script provides a guided example of using the parquet generation tools.

#### Usage

```bash
# Run from project root
./scripts/demo_parquet.sh
```

#### What it does

1. **Generates** a sample Parquet file with 50 records using a fixed seed for reproducible results
2. **Verifies** the generated file and displays the first 5 records
3. **Shows statistics** about the generated data
4. **Provides next steps** and cleanup instructions

#### Example Output

The script will:
- Create a `demo_data.parquet` file with sample geographic data
- Display file size and record information
- Show decoded WKB coordinates and statistics
- Suggest follow-up commands for further exploration

#### Customization

You can modify the script variables at the top:
- `OUTPUT_FILE`: Change the output filename
- `RECORD_COUNT`: Adjust number of records to generate
- `SEED`: Change random seed for different data patterns

## Adding New Scripts

When adding new scripts:

1. Use Go for scripts that need to import bbox packages
2. Use shell scripts for simple file operations
3. Add documentation to this README
4. Follow the naming convention: `verb_noun.go` or `verb_noun.sh`
5. Include `--help` flag support where appropriate