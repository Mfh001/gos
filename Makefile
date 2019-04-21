# Bootstrap
setup:
	sh dep_install.sh
	mkdir -p src/goslib/src/gen/proto
	mkdir -p src/goslib/src/gen/db
	protoc -I src/game/src/protos --go_out=src/goslib/src/gen/db src/game/src/protos/schema.proto
	cd src/goslib && protoc -I src/rpc_proto --go_out=plugins=grpc:src/gen/proto src/rpc_proto/*.proto
	cd generator && ./tools/gen_routes
	cd generator && ./tools/gen_protocol
	cd generator && ./tools/gen_excels

gen_player_schema:
	mkdir -p src/goslib/src/gen/db
	protoc -I src/game/src/protos --go_out=src/goslib/src/gen/db src/game/src/protos/schema.proto

# Install dependent go packages
dep_install:
	sh dep_install.sh

# Build apps: game, world
build:
	sh build_gos.sh

start_all:
	mkdir -p logs
	nohup ./game/bin/game > logs/game.log &
	nohup ./world/bin/world > logs/world.log &

# Generate framework grpc proto
gen_rpc_proto:
	mkdir -p src/goslib/src/gen/proto
	cd src/goslib && protoc -I src/rpc_proto --go_out=plugins=grpc:src/gen/proto src/rpc_proto/*.proto

# Generate client-server communicate protocol files
gen_protocol:
	cd generator && ./tools/gen_routes
	cd generator && ./tools/gen_protocol

# Generate config files from excels
gen_configs:
	cd generator && ./tools/gen_excels

######################## Docker ######################## 
build_ubuntu_apps:
	docker run --rm -v $(shell pwd):/usr/src/gos -w /usr/src/gos -e GOOS=linux -e GOARCH=amd64 golang:ubuntu sh build_gos.sh

build_docker_images:
	docker run --rm -v $(shell pwd):/usr/src/gos -w /usr/src/gos -e GOOS=linux -e GOARCH=amd64 golang:alpine sh build_gos.sh
	cd dockers && ./build-dockers

push_docker_images:
	docker tag gos-game-app savin198/gos-game-app
	docker tag gos-world-app savin198/gos-world-app
	docker push savin198/gos-game-app
	docker push savin198/gos-world-app

run_dockers:
	cd dockers && ./run-apps

load_docker_images:
	cd dockers && ./load-images

delete_k8s:
	kubectl delete -f dockers/k8s/deployments/world-service.yaml
	kubectl delete -f dockers/k8s/deployments/game-deployment.yaml

apply_k8s:
	kubectl apply -f dockers/k8s/deployments/world-service.yaml
	kubectl apply -f dockers/k8s/deployments/game-deployment.yaml
