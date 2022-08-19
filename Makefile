.DEFAULT_GOAL := build

fmt:
	go fmt ./...
.PHONY:fmt

lint: fmt
		golint ./...
.PHONY:fmt

vet: fmt
		go vet ./...
.PHONE:vet

build: vet
		go build hello.go
.PHONY:build

run: vet
		go run hello.go
.PHONY:run
