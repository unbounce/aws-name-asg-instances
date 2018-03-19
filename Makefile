.DEFAULT_GOAL := build
.PHONY: build test

project.repo := github.com/unbounce/aws-name-asg-instances

build:
	GOOS=linux GOARCH=amd64 go build -o .build/main $(project.repo)

test:
	go test $(project.repo)

