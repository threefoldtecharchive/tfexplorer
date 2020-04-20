OUT = $(shell realpath -m bin)
GOPATH := $(shell go env GOPATH)
branch = $(shell git symbolic-ref -q --short HEAD || git describe --tags --exact-match)
revision = $(shell git rev-parse HEAD)
dirty = $(shell test -n "`git diff --shortstat 2> /dev/null | tail -n1`" && echo "*")
version = github.com/threefoldtech/zos/pkg/version
ldflags = '-w -s -X $(version).Branch=$(branch) -X $(version).Revision=$(revision) -X $(version).Dirty=$(dirty)'

.PHONY: frontend server tfexplorer tffarmer tfuser

all: tfexplorer tffarmer

getdeps:
	@echo "Installing golint" && go install golang.org/x/lint/golint
	@echo "Installing gocyclo" && go install github.com/fzipp/gocyclo
	@echo "Installing misspell" && go install github.com/client9/misspell/cmd/misspell
	@echo "Installing ineffassign" && go install github.com/gordonklaus/ineffassign
	@echo "Installing statik" && go install github.com/rakyll/statik

verifiers: vet fmt lint cyclo spelling staticcheck

vet:
	@echo "Running $@"
	@go vet -atomic -bool -copylocks -nilfunc -printf -rangeloops -unreachable -unsafeptr -unusedresult ./...

fmt:
	@echo "Running $@"
	@gofmt -d $(shell ls **/*.go | grep -v statik)

lint:
	@echo "Running $@"
	golint -set_exit_status $(shell go list ./... | grep -v stubs | grep -v generated| grep -v migrations | grep -v statik)

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

# Builds minio, runs the verifiers then runs the tests.
check: test
test: verifiers
	# we already ran vet separately, so safe to turn it off here
	@echo "Running unit tests"
	for pkg in $(shell go list ./... ); do \
		go test -v -vet=off $$pkg; \
	done

testrace: verifiers
	@echo "Running unit tests with -race flag"
	# we already ran vet separately, so safe to turn it off here
	@CGO_ENABLED=1 go test -v -vet=off -race ./...

frontend:
	cd frontend && yarn install
	cd frontend && NODE_ENV=production yarn build

server:
	cd cmds/tffarmer && go generate
	cd cmds/tffarmer && go build -ldflags $(ldflags) -o $(OUT)/tfexplorer

tfexplorer: frontend server

tfuser:
	cd cmds/tffarmer && go build -ldflags $(ldflags) -o $(OUT)/tfuser

tffarmer:
	cd cmds/tffarmer && go build -ldflags $(ldflags) -o $(OUT)/tffarmer
	
