all: fmt lint vet client gate master notice store 

LDFLAGS += -X "github.com/dearcode/candy/util.BuildTime=$(shell date)"
LDFLAGS += -X "github.com/dearcode/candy/util.BuildVersion=$(shell git rev-parse HEAD)"

FILES := $$(find . -name '*.go' | grep -vE 'vendor') 
PACKAGES  := $$(go list ./...)
SOURCE_PATH := store master notice gate util

golint:
	go get github.com/golang/lint/golint  

godep:
	go get github.com/tools/godep

.PHONY: client gate master notice store



meta:
	@cd meta; make; cd ..; 

lint: golint
	@for path in $(SOURCE_PATH); do \
		echo "golint $$path" ; \
		golint $$path ; \
	done;

clean:
	@rm -rf bin

fmt: 
	@for path in $(SOURCE_PATH); do \
		echo "gofmt -s -l -w $$path" ; \
		gofmt -s -l -w $$path ; \
		echo "goimports -l -w $$path" ; \
		goimports -l -w $$path ; \
	done;

vet:
	go tool vet $(FILES) 2>&1
	go tool vet --shadow $(FILES) 2>&1


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

client: godep
	@echo "make client"
	@godep go build -ldflags '$(LDFLAGS)' -o bin/client ./candy.go

test:
	@for path in $(SOURCE_PATH); do \
		echo "go test ./$$path" ; \
		go test "./"$$path ; \
	done;


