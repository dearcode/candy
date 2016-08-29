all: godep lint meta gate master notice store 

GOLINT := golint

$(GOLINT):
	go get -u github.com/golang/lint/golint  

GODEP := godep

$(GODEP):
	go get github.com/tools/godep

.PHONY: meta gate master notice store


meta:
	@cd meta; make; cd ..; 

lint:
	$(GOLINT) gate/
	$(GOLINT) store/

clean:
	@rm -rf bin

fmt:
	gofmt -s -l -w .
	goimports -l -w .

vet:
	go tool vet . 2>&1
	go tool vet --shadow . 2>&1


gate:
	$(GODEP) go build -o bin/gate ./cmd/gate/main.go

master:
	$(GODEP) go build -o bin/master ./cmd/master/main.go

notice:
	$(GODEP) go build -o bin/notice ./cmd/notice/main.go

store:
	$(GODEP) go build -o bin/store ./cmd/store/main.go
