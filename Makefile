all: lint store master notifer gate tool client 

GitTime = github.com/dearcode/candy/util.GitTime
GitMessage = github.com/dearcode/candy/util.GitMessage

LDFLAGS += -X "$(GitTime)=$(shell git log --pretty=format:'%ct' -1)"
LDFLAGS += -X "$(GitMessage)=$(shell git log --pretty=format:'%cn %s %b' -1)"

FILES := $$(find . -name '*.go' | grep -vE 'vendor') 
SOURCE_PATH := store master notifer gate util

golint:
	go get github.com/golang/lint/golint  

godep:
	go get github.com/tools/godep

megacheck:
	go get honnef.co/go/tools/cmd/megacheck

meta:
	@cd meta; make; cd ..; 

lint: golint megacheck
	@for path in $(SOURCE_PATH); do golint $$path; done;
	@for path in $(SOURCE_PATH); do gofmt -s -l -w $$path;  done;
	go tool vet $(FILES) 2>&1
	megacheck ./...

clean:
	@rm -rf bin


store:godep
	godep go build -o bin/$@ -ldflags '$(LDFLAGS)' cmd/$@/main.go 
				     		                       			
gate:godep		     		                       			
	godep go build -o bin/$@ -ldflags '$(LDFLAGS)' cmd/$@/main.go 
				     		                       			
master:godep	     		                       			
	godep go build -o bin/$@ -ldflags '$(LDFLAGS)' cmd/$@/main.go 
				     		                       			
notifer:godep
	godep go build -o bin/$@ -ldflags '$(LDFLAGS)' cmd/$@/main.go 

client: godep
	godep go build -o bin/$@ -ldflags '$(LDFLAGS)' candy.go
                            
tool: godep                 
	godep go build -o bin/$@ -ldflags '$(LDFLAGS)' cmd/$@/main.go

test:
	@for path in $(SOURCE_PATH); do echo "go test ./$$path"; go test "./"$$path; done;


