```
go build
./bbox --draw
```

Unit Tests
```
go test -v ./...
```

Integration Tests
```
./integration_tests/setup.sh
bats integration_tests
```

# Usage

### Draw a bounding box in the browser
`bbox --draw`

### Draw a bounding box and output it in the terminal as GeoJSON
`bbox --draw --output geojson --geojson-indent 4`

### Create a bounding box from a center point
`bbox --center 1.0 2.0 --width 10 --height 10`

### Create a bounding box from a combination of sides and dimensions
```
bbox --left 1.0 --bottom 1.0 --right 2.0 --top 2.0
bbox --left 1.0 --bottom 1.0 --width 2.0 --height 2.0
bbox --left 1.0 --bottom 1.0 --width 2.0 --height 2.0
```

### Use distance unit in the dimensions (TODO)
`bbox --left 1.0 --bottom 2.0 --width 2.0mi --height 2.0mi`
units: mi,ft,km,m

### Create a boundng box from a geocoded place name
`bbox --place "Boston, MA"`

### Use custom geocoding services like Mapbox
```bash
# Using Mapbox Geocoding API v6 with access token
bbox --place "San Francisco, CA" \
  --geocoder-url "https://api.mapbox.com/search/geocode/v6/forward?q=%s&access_token=YOUR_MAPBOX_ACCESS_TOKEN"
```

### Accept input in a variery of formats
```
bbox --output wkt -- 1.0 1.0 2.0 2.0
bbox "POLYGON((1.0 1.0, 2.0 1.0, 2.0 2.0, 1.0 2.0, 1.0 1.0))" --output comma
```

### Accept input from stdin
```
cat whatevs.geojson | bbox --output wkt
```

### Create a bounding box from gis files
```
bbox whatevs.shp
bbox whatevs.geojson
bbox whatevs.geojsonl
bbox whatevs.osm
```

### specify a bbox on the cli -- then edit it in the browser
`bbox --center 1.0 2.0 --width 10 --height 10 --draw`

### center - get the center of the box
`bbox center -- 1.0 1.0 2.0 2.0`

### slice - Slice the bounding box into smaller boxes
`bbox slice --center 1.0 2.0 --width 10 --height 10 --rows 5 --columns 10`

### Tile (TODO)
`bbox tile --center 1.0 2.0 --width 10 --height 10`
TODO way to limit the tiles

### API (TODO)
`bbox serve-api`


Output formats:
```
-o comma
-o space
-o tab
-o wkt
-o hexwkb
-o geojson
-o overpass-ql # TODO
-o url=osm
-o "go-template={{.Left}} {{.Bottom}} {{.Right}} {{.Top}}"
```

# TODO
* add support wkt/wkb/hexwkb formats to ParseData (and give better name)
* align input and output options across commands
    * dublinecore output for center/collection
* Cleanup internal errors vs user errors
* basic projection handling
    * read the projection if we can
    * for projections that are close WGS84, allow --drow but show a warning
    * for those that aren't show error
    * read proj from:
        * WKB
        * geoparquet
        * shapefile
    * output CRS:
        * EWKT
        * WKB
        * DublinCore
    * handle case where files have mismatched SRIDs
* cleanup draw UI
    * Important: Show popup success message with button to close window when done - disable/blank out UI
    * touch interactions/small screen UI
    * preview bbox in common formats
    * UI mode that just opens the UI in the browser
* output formats
    * lines
    * overpass ql
    * r/sf format
* option to specify decimal precision/format
* handle projections
    * https://github.com/twpayne/go-proj
    * income port(?) active 4 months ago https://github.com/go-spatial/proj
    * Map tiler api
    * call out to proj
    * implement basic projections
    * https://github.com/wroge/wgs84
* units on width, height, buffer options
* clean input error messaging
* add http api
* Text description of Bbox - get closest major city to all four corners and center, and the dedup to describe
“12km x 12km box 45 km north east of Minneapolis

* tile command
* area command?
