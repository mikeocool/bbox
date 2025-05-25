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
