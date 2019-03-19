#!/usr/bin/env bash
basePath=$(pwd)/src/vendor

# Build auth
export GOPATH=$basePath:$(pwd)/src/auth:$(pwd)/src/goslib
go install auth

# Build agent
export GOPATH=$basePath:$(pwd)/src/agent:$(pwd)/src/goslib:$(pwd)/src/game
go install agent

# Build game
export GOPATH=$basePath:$(pwd)/src/game:$(pwd)/src/goslib
go install game

# Build world
export GOPATH=$basePath:$(pwd)/src/world:$(pwd)/src/goslib
go install world
