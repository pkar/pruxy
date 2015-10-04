REPO              = github.com/pkar/pruxy
APP               = pruxy
CMD               = $(REPO)/cmd/$(APP)
IMAGE_NAME        = pkar/$(APP)
IMAGE_TAG         = latest
IMAGE_SPEC        = $(IMAGE_NAME):$(IMAGE_TAG)
UNAME             := $(shell uname | awk '{print tolower($0)}')
TAG               = v0.0.1

vendor:
	go get -u github.com/coreos/go-etcd/etcd

build_docker:
	docker build -t $(IMAGE_SPEC) .

build_linux:
	mkdir -p bin/linux_amd64
	GOARCH=amd64 GOOS=linux go build -o bin/linux_amd64/$(APP) ./cmd/$(APP)/main.go

build_darwin:
	mkdir -p bin/darwin_amd64
	go build -o bin/darwin_amd64/$(APP) ./cmd/$(APP)/main.go

build:
	$(MAKE) build_$(UNAME)

release:
	$(MAKE) build
	cd bin/$(UNAME)_amd64 && tar -czvf runit-$(TAG).$(UNAME).tar.gz runit
	mv bin/$(UNAME)_amd64/runit-$(TAG).$(UNAME).tar.gz bin/

install:
	go install $(CMD)

run:
	go run cmd/$(APP)/main.go

test:
	go test -cover .

testv:
	go test -v -cover .

testf:
	# make testf TEST=TestRunCmd
	go test -v -test.run="$(TEST)"

testrace:
	go test -race .

bench:
	go test ./... -bench=.

vet:
	go vet ./...

coverprofile:
	# run tests and create coverage profile
	go test -coverprofile=coverage.out .
	# check heatmap
	go tool cover -html=coverage.out

.PHONY: vendor test install release build
