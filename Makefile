all: golint meta gate master notice store 

.PHONY: meta gate master notice store

meta:
	 @cd meta; make; cd ..; 

golint:
	golint gate/

clean:
	@rm -rf bin

fmt:
	gofmt -s -l -w .
	goimports -l -w .

vet:
	go tool vet . 2>&1
	go tool vet --shadow . 2>&1


gate:
	godep go build -o bin/gate ./cmd/gate/main.go

master:
	godep go build -o bin/master ./cmd/master/main.go

notice:
	godep go build -o bin/notice ./cmd/notice/main.go

store:
	godep go build -o bin/store ./cmd/store/main.go
