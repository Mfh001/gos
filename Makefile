all:
	./tools/gen_protocol
	./tools/gen_routes
	go install server

start:
	./bin/server
