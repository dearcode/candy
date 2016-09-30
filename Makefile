all: lint client gate master notice store 

LDFLAGS += -X "github.com/dearcode/candy/util.BuildTime=$(shell date)"
LDFLAGS += -X "github.com/dearcode/candy/util.BuildVersion=$(shell git rev-parse HEAD)"

golint:
	go get github.com/golang/lint/golint  

godep:
	go get github.com/tools/godep

.PHONY: client gate master notice store



meta:
	@cd meta; make; cd ..; 

lint: golint
	golint gate/
	golint store/
	golint client/
	golint master/
	golint notice/
	golint util/

clean:
	@rm -rf bin

fmt:
	gofmt -s -l -w .
	goimports -l -w .

vet:
	go tool vet . 2>&1
	go tool vet --shadow . 2>&1


gate: godep
	@go tool vet ./gate/ 2>&1
	@echo "make gate"
	@godep go build -ldflags '$(LDFLAGS)' -o bin/gate ./cmd/gate/main.go

master: godep
	@echo "make master"
	@godep go build -ldflags '$(LDFLAGS)' -o bin/master ./cmd/master/main.go

notice: godep
	@echo "make notice"
	@godep go build -ldflags '$(LDFLAGS)' -o bin/notice ./cmd/notice/main.go

store: godep
	@echo "make store"
	@godep go build -ldflags '$(LDFLAGS)' -o bin/store ./cmd/store/main.go

client: godep
	@echo "make client"
	@godep go build -ldflags '$(LDFLAGS)' -o bin/client ./candy.go

test:
	@go test ./client/
	@go test ./notice/
	@go test ./store/

