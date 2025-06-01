setup() {
    load 'libs/bats-support/load'
    load 'libs/bats-assert/load'

    # get the containing directory of this file
    # use $BATS_TEST_FILENAME instead of ${BASH_SOURCE[0]} or $0,
    # as those will point to the bats executable's location or the preprocessed file respectively
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )" >/dev/null 2>&1 && pwd )"
}

@test "raw space input" {
    run ./bbox -- -91.020128093 48.499043895 -90.586309459 48.691033039
    assert_output "-91.020128093 48.499043895 -90.586309459 48.691033039"
    assert_success
}

@test "bounds input" {
    run ./bbox -l -91.020128093 -b 48.499043895 -r -90.586309459 -t 48.691033039
    assert_output "-91.020128093 48.499043895 -90.586309459 48.691033039"
    assert_success
}

@test "stdin spaces input" {
    run /bin/bash -c "echo '-91.020128093 48.499043895 -90.586309459 48.691033039' | ./bbox"
    assert_output "-91.020128093 48.499043895 -90.586309459 48.691033039"
    assert_success
}

@test "stdin newlines input" {
    run /bin/bash -c "echo -e '1.0 1.0\n2.0 4.0\n3.0 1.0\n' | ./bbox"
    assert_output "1 1 3 4"
    assert_success
}

@test "stdin geojson file" {
    run /bin/bash -c "cat $DIR/data/campsites.geojson | ./bbox"
    assert_output "-92.42919378022346 47.77639791033817 -90.03548429130946 48.35501085637799"
    assert_success
}

@test "invalid stdin" {
    run /bin/bash -c "echo '' | ./bbox"
    assert_output --partial "invalid input"
    assert_failure
}

@test "go-template output" {
    run ./bbox -l -1 --bottom -2 -r 1 -t 2  -o go-template='{{.Top}} {{.Right}} {{.Bottom}} {{.Left}}'
    assert_output "2 1 -2 -1"
    assert_success
}

@test "load geojson" {
    run ./bbox --file $DIR/data/campsites.geojson
    assert_output "-92.42919378022346 47.77639791033817 -90.03548429130946 48.35501085637799"
    assert_success
}

@test "load geojson without extension" {
    run ./bbox --file $DIR/data/coords
    assert_output "-10 -5 10 5"
    assert_success
}

@test "load shapfile" {
    run ./bbox --file $DIR/data/ne_10m_populated_places_simple/ne_10m_populated_places_simple.shp
    assert_output "-179.5899789 -89.9999998 179.3833036 82.4833232"
    assert_success
}

@test "pipe geojson" {
    run /bin/bash -c "cat $DIR/data/campsites.geojson | ./bbox"
    assert_output "-92.42919378022346 47.77639791033817 -90.03548429130946 48.35501085637799"
    assert_success
}

@test "pipe shapefile" {
    run /bin/bash -c "cat $DIR/data/ne_10m_populated_places_simple/ne_10m_populated_places_simple.shp | ./bbox"
    assert_output "-179.5899789 -89.9999998 179.3833036 82.4833232"
    assert_success
}

@test "simple center" {
    run ./bbox center 10 17 20 20 -o comma
    assert_output "15,18.5"
    assert_success
}
