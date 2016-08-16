all: sync golint meta gate

.PHONY: meta gate

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

