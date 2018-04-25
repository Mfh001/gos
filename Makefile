setup:
	cd GosLib && protoc -I src/gosRpcProto --go_out=plugins=grpc:src/gosRpcProto src/gosRpcProto/*.proto
	cd GameApp && ./tools/gen_routes
	cd GameApp && ./tools/gen_protocol
	cd GameApp && bundle exec rake generate_tables

rpc_proto:
	cd GosLib && protoc -I src/gosRpcProto --go_out=plugins=grpc:src/gosRpcProto src/gosRpcProto/*.proto

tcp_protocol:
	cd GameApp && ./tools/gen_routes
	cd GameApp && ./tools/gen_protocol

generate_tables:
	cd GameApp && bundle exec rake generate_tables

gopath:
	export GOPATH=$HOME/.go:$(pwd)/AuthApp:$(pwd)/ConnectApp:$(pwd)/GameApp:$(pwd)/WorldApp:$(pwd)/GosLib
	export PATH=$PATH:$GOPATH/bin

build:
	sh build_gos.sh

build_linux:
	sudo docker run --rm -v $(shell pwd):/usr/src/gos -w /usr/src/gos -e GOOS=linux -e GOARCH=amd64 golang:latest make build
