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
