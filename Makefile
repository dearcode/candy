all: sync golint meta gate master notice store 

.PHONY: meta gate master notice store

GLOCK := glock

$(GLOCK): 
	go get github.com/robfig/glock

sync: $(GLOCK)
	$(GLOCK) sync github.com/dearcode/candy/

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
	go build -o bin/gate ./cmd/gate/main.go

master:
	go build -o bin/master ./cmd/master/main.go

notice:
	go build -o bin/notice ./cmd/notice/main.go

store:
	go build -o bin/store ./cmd/store/main.go
