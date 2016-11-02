all: fmt lint vet store master notice gate tool client 

LDFLAGS += -X "github.com/dearcode/candy/util.BuildTime=$(shell date -R)"
LDFLAGS += -X "github.com/dearcode/candy/util.BuildVersion=$(shell git rev-parse HEAD)"

FILES := $$(find . -name '*.go' | grep -vE 'vendor') 
SOURCE_PATH := store master notice gate util

golint:
	go get github.com/golang/lint/golint  

godep:
	go get github.com/tools/godep


meta:
	@cd meta; make; cd ..; 

lint: golint
	@for path in $(SOURCE_PATH); do echo "golint $$path"; golint $$path; done;

clean:
	@rm -rf bin

fmt: 
	@for path in $(SOURCE_PATH); do echo "gofmt -s -l -w $$path";  gofmt -s -l -w $$path;  done;

vet:
	go tool vet $(FILES) 2>&1
	go tool vet --shadow $(FILES) 2>&1

store:godep
	godep go build -o bin/$@ -ldflags '$(LDFLAGS)' cmd/$@/main.go 
				     		                       			
gate:godep		     		                       			
	godep go build -o bin/$@ -ldflags '$(LDFLAGS)' cmd/$@/main.go 
				     		                       			
master:godep	     		                       			
	godep go build -o bin/$@ -ldflags '$(LDFLAGS)' cmd/$@/main.go 
				     		                       			
notice:godep	     		                       			
	godep go build -o bin/$@ -ldflags '$(LDFLAGS)' cmd/$@/main.go 

client: godep
	godep go build -o bin/$@ -ldflags '$(LDFLAGS)' candy.go
                            
tool: godep                 
	godep go build -o bin/$@ -ldflags '$(LDFLAGS)' cmd/$@/main.go

test:
	@for path in $(SOURCE_PATH); do echo "go test ./$$path"; go test "./"$$path; done;


