all: pb

.PHONY: pb

pb: message.proto
	protoc --go_out=plugins=grpc:. $+

