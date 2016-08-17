all: pb

.PHONY: pb

pb: common.proto gate.proto master.proto notice.proto store.proto
	protoc --go_out=plugins=grpc:. $+

