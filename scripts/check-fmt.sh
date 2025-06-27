#!/bin/bash

set -e

echo "Checking go fmt..."

UNFORMATTED=$(gofmt -s -l .)

if [ -n "$UNFORMATTED" ]; then
    echo "The following files are not formatted:"
    echo "$UNFORMATTED"
    echo ""
    echo "Run 'go fmt ./...' to fix formatting issues."
    exit 1
fi

echo "All files are properly formatted."