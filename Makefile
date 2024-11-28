PACKAGES := $(shell go list ./...)
name := $(shell basename ${PWD})

all: help

.PHONY: help
help: Makefile
	@echo
	@echo " Choose a make command to run"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## init: initialize project (make init module=github.com/user/project)
.PHONY: init
init:
	go mod init ${module}
	go install github.com/cosmtrek/air@latest
	asdf reshim golang

## vet: vet code
.PHONY: vet
vet:
	go vet $(PACKAGES)

## test: run unit tests
.PHONY: test
test:
	go test -race -cover $(PACKAGES)

## build: build a binary
.PHONY: build
build:
	go build -v -o ./tmp/main .

## docker-build: build project into a docker container image
.PHONY: docker-build
docker-build: test
	GOPROXY=direct docker buildx build -t ${name} .

## docker-run: run project in a container
.PHONY: docker-run
docker-run:
	docker run -it --rm -p 8080:8080 ${name}

## start: build and run local project
.PHONY: start
start:
	air

## new-migration: create new tern migration
.PHONY: new-migration
new-migration:
	TERN_MIGRATIONS=sqlc/migrations tern new migration ${NAME}

## db-migrate: run tern migrations
.PHONY: db-migrate
db-migrate:
	TERN_MIGRATIONS=sqlc/migrations tern migrate

## sqlc-generate: generate sqlc files
.PHONY: sqlc-generate
sqlc-generate:
	sqlc generate

## proto-generate: generate protobuf files
.PHONY: proto-generate
proto-generate:
	 go get github.com/hyperremix/song-contest-rater-proto@$(VERSION) && protoc -I=${GOPATH}/pkg/mod/github.com/hyperremix/song-contest-rater-proto@$(VERSION) --go_out=. --go-grpc_out=. ${GOPATH}/pkg/mod/github.com/hyperremix/song-contest-rater-proto@$(VERSION)/*.proto && go mod tidy
