-include .env
export $(shell sed 's/=.*//' .env)

PACKAGES := $(shell go list ./...)
name := $(shell basename ${PWD})
SONGCONTESTRATERSERVICE_PROTO_VERSION := v1.0.34

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

## env: output environment variables
.PHONY: env
env:
	(env | grep '^SONGCONTESTRATERSERVICE') 2> /dev/null

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
	GOPROXY=direct docker build -t ${name} .

## docker-run: run project in a container
.PHONY: docker-run
docker-run:
	docker run -it --rm -p 8080:8080 --env-file .env ${name}

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
	go get github.com/hyperremix/song-contest-rater-proto@$(SONGCONTESTRATERSERVICE_PROTO_VERSION) && protoc -I=$$(go env GOPATH)/pkg/mod/github.com/hyperremix/song-contest-rater-proto@$(SONGCONTESTRATERSERVICE_PROTO_VERSION) --go_out=. --go-grpc_out=. $$(go env GOPATH)/pkg/mod/github.com/hyperremix/song-contest-rater-proto@$(SONGCONTESTRATERSERVICE_PROTO_VERSION)/*.proto && go mod tidy

## fly-db-proxy: start fly.io db proxy
.PHONY: fly-db-proxy
fly-db-proxy:
	fly proxy 5434:5432 -a song-contest-rater-service-db

## deploy: deploy to fly.io
.PHONY: deploy
deploy:
	fly deploy