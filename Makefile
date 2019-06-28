#The MIT License (MIT)
#
#Copyright (c) 2018 SavinMax. All rights reserved.
#
#Permission is hereby granted, free of charge, to any person obtaining a copy
#of this software and associated documentation files (the "Software"), to deal
#in the Software without restriction, including without limitation the rights
#to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
#copies of the Software, and to permit persons to whom the Software is
#furnished to do so, subject to the following conditions:
#
#The above copyright notice and this permission notice shall be included in
#all copies or substantial portions of the Software.
#
#THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
#IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
#AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
#LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
#OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
#THE SOFTWARE.

# Bootstrap
setup:
	cd src/game && go mod dity
	cd src/goslib && go mod dity
	cd src/world && go mod dity
	mkdir -p src/goslib/gen/db
	protoc -I src/game/protos --go_out=src/goslib/gen/db src/game/protos/schema.proto
	mkdir -p src/goslib/gen/proto
	cd src/goslib && protoc -I rpc_proto --go_out=plugins=grpc:gen/proto rpc_proto/*.proto
	cd generator && ./tools/gen_routes
	cd generator && ./tools/gen_protocol
	cd generator && ./tools/gen_excels

gen_player_schema:
	mkdir -p src/goslib/gen/db
	protoc -I src/game/protos --go_out=src/goslib/gen/db src/game/protos/schema.proto

# Install dependent go packages
dep_install:
	sh dep_install.sh

# Build apps: game, world
build:
	sh build_gos.sh

build_linux:
	sh build_gos_linux.sh

# Generate framework grpc proto
gen_rpc_proto:
	mkdir -p src/goslib/gen/proto
	cd src/goslib && protoc -I rpc_proto --go_out=plugins=grpc:gen/proto rpc_proto/*.proto

# Generate client-server communicate protocol files
gen_protocol:
	cd generator && ./tools/gen_routes
	cd generator && ./tools/gen_protocol

# Generate config files from excels
gen_configs:
	cd generator && ./tools/gen_excels

load_config:
	cd generator && redis-cli -x SET __gs_configs__  < configData.json.gz

######################## Docker ######################## 
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

publish_dev_game:
	sh build_gos_linux.sh
	sup dev deploy-game

setup_dev_game:
	sh build_gos_linux.sh
	sup dev setup

deploy_dev_game:
	sh build_gos_linux.sh
	sup dev deploy
