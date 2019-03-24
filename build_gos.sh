#!/usr/bin/env bash
basePath=$(pwd)/src/vendor:$(pwd)/src/goslib

# Build game
export GOPATH=$basePath:$(pwd)/src/game
go install game

# Build world
export GOPATH=$basePath:$(pwd)/src/world
go install world
