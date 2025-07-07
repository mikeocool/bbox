# Shapefile CRS Implementation

## Original Request

Implement reading the projection from a shapefile, using it to figure out the SRID and set that on the bbox that is returned from the LoadShapefile in `input/shapefile.go`.

## Full Plan for Shapefile CRS Implementation

### Current State Analysis
- `core.Bbox` already has `Srid` field support
- WKT parser exists but handles **geometry WKT** (POINT, POLYGON, etc.), not **projection WKT**
- Shapefile CRS info is stored in `.prj` files containing projection WKT
- Most `.prj` files don't include `AUTHORITY` clauses

### Implementation Steps

1. **Modify `LoadShapefile` function** (`input/shapefile.go:37-44`):
   - Look for corresponding `.prj` file (same basename, `.prj` extension)
   - Read and parse the projection WKT to extract SRID
   - Set the SRID on the returned bbox

2. **Create projection WKT parser** (new function in `input/shapefile.go`):
   - **Primary approach**: Look for `AUTHORITY["EPSG","xxxx"]` patterns
   - **Fallback approach**: Use lookup table for common projection names/parameters
   - Return `UnknownCrs` if no match found

3. **Implement projection lookup table**:
   - Map common projection names to EPSG codes:
     - `GCS_WGS_1984` → EPSG:4326
     - `NAD_1983_UTM_Zone_15N` → EPSG:26915  
     - `WGS_1984_Web_Mercator` → EPSG:3857
     - `GCS_North_American_1983` → EPSG:4269
     - Geographic coordinate systems vs projected systems

4. **Error handling**:
   - Handle missing `.prj` files gracefully (default to `UnknownCrs`)
   - Handle malformed projection WKT
   - Preserve existing shapefile parsing behavior

### Technical Implementation Details

#### Function Structure
```go
func parseProjectionWKT(prjContent string) int {
    // 1. Try to extract AUTHORITY["EPSG","xxxx"] 
    // 2. Fall back to projection name lookup
    // 3. Return UnknownCrs if no match
}

func LoadShapefile(filename string) (core.Bbox, error) {
    // Existing shapefile parsing...
    bbox := // ... current parsing logic
    
    // Look for .prj file
    prjFile := strings.TrimSuffix(filename, ".shp") + ".prj"
    srid := parseProjectionFile(prjFile)
    bbox.Srid = srid
    
    return bbox, nil
}
```

#### Parsing Strategy
1. **AUTHORITY parsing**: Regex or string parsing for `AUTHORITY["EPSG","(\d+)"]`
2. **Name-based lookup**: Extract projection name from `PROJCS["name",...` or `GEOGCS["name",...`
3. **Parameter-based matching**: For UTM zones, match zone numbers and hemispheres

#### Lookup Table Categories
- **Geographic (lat/lon)**: WGS84, NAD83, NAD27
- **UTM projections**: All zones for WGS84, NAD83
- **State plane**: Common US state plane systems  
- **Web mapping**: Web Mercator, Google projections
- **National grids**: British National Grid, etc.

### Key Design Decisions

- **File handling**: Use `filepath.Join` and `strings.TrimSuffix` for cross-platform compatibility
- **Backward compatibility**: Maintain existing function signatures
- **Error tolerance**: Missing `.prj` files don't cause parsing failures
- **Performance**: Simple string matching before complex parsing
- **Extensibility**: Easy to add new projection mappings

### Testing Strategy
- Test with `.prj` files that have `AUTHORITY` clauses
- Test with common `.prj` files without `AUTHORITY` clauses  
- Test with missing `.prj` files
- Test with malformed `.prj` files
- Verify SRID propagation through the bbox pipeline

This approach handles both modern shapefiles (with AUTHORITY) and legacy shapefiles (without AUTHORITY) while maintaining robust error handling.