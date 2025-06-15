```
go build
./bbox draw
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

CLI examples:

# Basics

### Draw a bounding box in the browser
bbox --draw

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
`bbox --file whatevs.shp`

### specify a bbox on the cli -- but then edit it in the browser
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
-o comma
-o space
-o wkt
-o hexwkb
-o geojson
-o overpass-ql
-o osm-url
-o "template={{.MinX}} {{.MinY}}"
-o kml

# TODO
* geojsonl -- input/output
* --place input
    * https://photon.komoot.io
* osm file input
    * https://github.com/paulmach/osm
* align input and output options across commands
* add github actions for testing
* cleanup draw UI
    * handle click and drag when creating box
    * Show popup success message with button to close window when done
    * touch interactions/small screen UI
    * vector tiles(?)
    * preview bbox in common formats
    * allow changing labels left/bottom/top/right, minx..., min lat, west/south/east/north
* Add a verbose flag
* kml input
* output formats
    * lines
    * overpass ql
    * option to open browser for url formats
* match input and output formats as closely as possible
* option to specify decimal precision/format
* handle projections
    * https://github.com/twpayne/go-proj
    * income port(?) active 4 months ago https://github.com/go-spatial/proj
    * Map tiler api
    * call out to proj
    * implement basic projections
* units on width, height, buffer options
* clean input error messaging
* add http api
* Text description of Bbox - get closest major city to all four corners and center, and the dedup to describe
“12km x 12km box 45 km north east of Minneapolis

* tile command
* area command?
