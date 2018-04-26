#!/usr/bin/env bash
basePath=$(pwd)/vendor

# Build auth
export GOPATH=$basePath:$(pwd)/auth:$(pwd)/goslib
go install auth

# Build agent
export GOPATH=$basePath:$(pwd)/agent:$(pwd)/goslib:$(pwd)/game
go install agent

# Build game
export GOPATH=$basePath:$(pwd)/game:$(pwd)/goslib
go install game

# Build world
export GOPATH=$basePath:$(pwd)/world:$(pwd)/goslib
go install world
