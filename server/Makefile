all: lint gate master notice store 

LDFLAGS += -X "github.com/dearcode/candy/server/util.BUILD_TIME=$(shell date +%s)"
LDFLAGS += -X "github.com/dearcode/candy/server/util.BUILD_VERSION=$(shell git rev-parse HEAD)"

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

test:
	go test ./android/
	go test ./store/

