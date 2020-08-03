root = $(shell pwd)
OUT = $(shell realpath -m bin)
GOPATH := $(shell go env GOPATH)
branch = $(shell git symbolic-ref -q --short HEAD || git describe --tags --exact-match)
revision = $(shell git rev-parse HEAD)
dirty = $(shell test -n "`git diff --shortstat 2> /dev/null | tail -n1`" && echo "*")
version = github.com/threefoldtech/zos/pkg/version
ldflags = '-w -s -X $(version).Branch=$(branch) -X $(version).Revision=$(revision) -X $(version).Dirty=$(dirty)'

.PHONY: frontend server tfexplorer tffarmer tfuser docs

all: tfexplorer tffarmer

getdeps:
	@echo "Installing golint" && go install golang.org/x/lint/golint
	@echo "Installing gocyclo" && go install github.com/fzipp/gocyclo
	@echo "Installing misspell" && go install github.com/client9/misspell/cmd/misspell
	@echo "Installing ineffassign" && go install github.com/gordonklaus/ineffassign
	@echo "Installing statik" && go install github.com/rakyll/statik
	@echo "Installing open api generator" && sudo npm install @openapitools/openapi-generator-cli -g

verifiers: vet fmt lint cyclo spelling staticcheck

vet:
	@echo "Running $@"
	@go vet -atomic -bool -copylocks -nilfunc -printf -rangeloops -unreachable -unsafeptr -unusedresult $(shell go list ./... | grep -v models| grep -v statik | grep -v tfuser|  grep -v migrations)

fmt:
	@echo "Running $@"
	@gofmt -d $(shell ls **/*.go | grep -v statik)

lint:
	@echo "Running $@"
	golint -set_exit_status $(shell go list ./... | grep -v models| grep -v statik | grep -v tfuser|  grep -v migrations)

ineffassign:
	@echo "Running $@"
	ineffassign .

cyclo:
	@echo "Running $@"
	gocyclo -over 100 .


spelling:
	misspell -i monitord -error $(shell ls **/*.go | grep -v statik)

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck -- ./...

check: test
test: verifiers
	# we already ran vet separately, so safe to turn it off here
	@CGO_ENABLED=1 go test -v -vet=off ./...

testrace: verifiers
	@echo "Running unit tests with -race flag"
	# we already ran vet separately, so safe to turn it off here
	@CGO_ENABLED=1 go test -v -vet=off -race ./...

frontend: frontend/src/*
	cd frontend && yarn install
	cd frontend && NODE_ENV=production yarn build --dest ../dist

server:
	cd cmds/tfexplorer && go generate
	cd cmds/tfexplorer && go build -ldflags $(ldflags) -o $(OUT)/tfexplorer

tfexplorer: frontend server

tfuser:
	cd cmds/tfuser && go build -ldflags $(ldflags) -o $(OUT)/tfuser

tffarmer:
	cd cmds/tffarmer && go build -ldflags $(ldflags) -o $(OUT)/tffarmer
	
stellar:
	cd cmds/stellar && go build -ldflags $(ldflags) -o $(OUT)/stellar

docs:
	cd docs && openapi-generator generate -i openapi.json -g html2 
	mv docs/index.html frontend/public/docs_legacy.html
	cd docs && openapi-generator generate -i api_v1.json -g html2
	mv docs/index.html frontend/public/docs_api_v1.html
	rm -rf docs/.openapi-generator
	rm -f docs/.openapi-generator-ignore

clean:
	rm -rf dist statik bin/*