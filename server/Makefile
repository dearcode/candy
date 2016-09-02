all: lint gate master notice store 

golint:
	go get github.com/golang/lint/golint  

godep:
	go get github.com/tools/godep

.PHONY: gate master notice store


meta:
	@cd meta; make; cd ..; 

lint: golint
	golint gate/
	golint store/

clean:
	@rm -rf bin

fmt:
	gofmt -s -l -w .
	goimports -l -w .

vet:
	go tool vet . 2>&1
	go tool vet --shadow . 2>&1


gate: godep
	godep go build -o bin/gate ./cmd/gate/main.go

master: godep
	godep go build -o bin/master ./cmd/master/main.go

notice: godep
	godep go build -o bin/notice ./cmd/notice/main.go

store: godep
	godep go build -o bin/store ./cmd/store/main.go
