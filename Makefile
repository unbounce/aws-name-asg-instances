.DEFAULT_GOAL := build
.PHONY: build test

project.repo := github.com/unbounce/aws-name-asg-instances

build:
	go build -o .build/main $(project.repo)

test:
	go test $(project.repo)

