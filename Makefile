dep_install:
	sh dep_install.sh

setup:
	cd goslib && protoc -I src/gos_rpc_proto --go_out=plugins=grpc:src/gos_rpc_proto src/gos_rpc_proto/*.proto
	cd game && ./tools/gen_routes
	cd game && ./tools/gen_protocol
	cd game && bundle exec rake generate_tables

build:
	sh build_gos.sh

start_all:
	mkdir -p logs
	nohup ./auth/bin/auth > logs/auth.log &
	nohup ./agent/bin/agent > logs/agent.log &
	nohup ./game/bin/game > logs/game.log &
	nohup ./world/bin/world > logs/world.log &

rpc_proto:
	cd goslib && protoc -I src/gos_rpc_proto --go_out=plugins=grpc:src/gos_rpc_proto src/gos_rpc_proto/*.proto

tcp_protocol:
	cd game && ./tools/gen_routes
	cd game && ./tools/gen_protocol

generate_tables:
	cd game && bundle exec rake generate_tables

build_ubuntu_apps:
	docker run --rm -v $(shell pwd):/usr/src/gos -w /usr/src/gos -e GOOS=linux -e GOARCH=amd64 golang:ubuntu sh build_gos.sh

build_docker_images:
	docker run --rm -v $(shell pwd):/usr/src/gos -w /usr/src/gos -e GOOS=linux -e GOARCH=amd64 golang:alpine sh build_gos.sh
	cd dockers && ./build-dockers

run_dockers:
	cd dockers && ./run-apps

load_docker_images:
	cd dockers && ./load-images
