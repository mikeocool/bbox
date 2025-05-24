```
go build
./bbox draw
```

```
go test -v ./...
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

# TODO
* draw with bbox populated from cli
* raw input
* raw input from stdin
* place input
* file input
* cleanup draw UI
* units on width and heightw
* center + area command
* handle projections
* clean input error messaging
* tile command
* split command
