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
```
# Basics
bbox --draw
bbox --draw --output comma

bbox --center 1.0 2.0 --width 10 --height 10

bbox --min-x 1.0 --min-y 1.0 --max-x 2.0 --max-y 2.0
bbox --min-x 1.0 --min-y 1.0 --width 2.0 --height 2.0
bbox --min-y 1.0 --max-x 2.0 --width 2.0 --height 2.0

bbox --min-y 1.0 --max-x 2.0 --width 2.0mi --height 2.0mi
units: mi,ft,km,m

bbox --place "Boston, MA" --width 10 --height 10

# open browser to get location (or --my-location ip ? and use api?)
bbox --my-location --width 10 --height 10

bbox --raw "1.0 1.0 2.0 2.0" --output wkt
bbox --raw "POLYGON((1.0 1.0, 2.0 1.0, 2.0 2.0, 1.0 2.0, 1.0 1.0))" --output comma

cat whatevs.txt | bbox --output wkt

bbox --file whatevs.shp

# specify a bbox on the cli -- but then open the browser and let you edit it
bbox --center 1.0 2.0 --width 10 --height 10 --draw

# Verbose (better name? -- summary?)
bbox verbose --center 1.0 2.0 --width 10 --height 10

bbox center
bbox area

# Slice
bbox slice --center 1.0 2.0 --width 10 --height 10 --x-slices 5 --y-slices 10

# Tile
bbox tile --center 1.0 2.0 --width 10 --height 10
# TODO way to limit the tiles

# API
bbox serve-api
```

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
* split command
* cleanup draw UI
    * vector tiles(?)
    * preview bbox in common formats
    * allow changing labels left/bottom/top/right, minx..., min lat, west/south/east/north
    * Show popup success message with button to close window when done
* Add a verbose flag
* Fix issue where parser raw ends up parsing everything as float
* geojsonl
* osm file input
    * https://github.com/paulmach/osm
* kml input
* output formats
    * lines
    * hex wkb
    * overpass ql
    * dublin core
    * option to open browser for url formats
    * variations on geojson
* place input
* tile command
* units on width and heightw
* option to specify decimal precision/format
* handle projections
    * https://github.com/twpayne/go-proj
    * income port(?) active 4 months ago https://github.com/go-spatial/proj
    * Map tiler api
    * call out to proj
    * implement basic projections
* clean input error messaging
* add http api
* --grow/shrink/padding args?
* area command?
* Text description of Bbox - get closest major city to all four corners and center, and the dedup to describe
“12km x 12km box 45 km north east of Minneapolis
