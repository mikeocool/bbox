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
