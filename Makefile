proto:
	cd GosLib && protoc -I src/gosRpcProto --go_out=plugins=grpc:src/gosRpcProto src/gosRpcProto/*.proto
