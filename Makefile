all: fmt lint vet cmd client

LDFLAGS += -X "github.com/dearcode/candy/util.BuildTime=$(shell date -R)"
LDFLAGS += -X "github.com/dearcode/candy/util.BuildVersion=$(shell git rev-parse HEAD)"

FILES := $$(find . -name '*.go' | grep -vE 'vendor') 
SOURCE_PATH := store master notice gate util
cmd := store master notice gate tool

golint:
	go get github.com/golang/lint/golint  

godep:
	go get github.com/tools/godep


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
	done;

vet:
	go tool vet $(FILES) 2>&1
	go tool vet --shadow $(FILES) 2>&1

cmd:godep
	@for cmd in $(cmd); do \
		echo "godep go build -ldflags '$(LDFLAGS)' -o bin/$$cmd ./cmd/$$cmd/main.go" ; \
		godep go build -ldflags '$(LDFLAGS)' -o bin/$$cmd ./cmd/$$cmd/main.go ; \
	done;

client: godep
	godep go build -ldflags '$(LDFLAGS)' -o bin/client ./candy.go

test:
	@for path in $(SOURCE_PATH); do \
		echo "go test ./$$path" ; \
		go test "./"$$path ; \
	done;


