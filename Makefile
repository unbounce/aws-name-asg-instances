.DEFAULT_GOAL := help
.PHONY: build test help deploy-iam-stack deploy-event-stack require-profile require-region check-iam-stack check-event-stack prepare-event-stack deploy-code

pwd := $(shell pwd)

project.name := aws-name-asg-instances
project.repo := github.com/unbounce/$(project.name)

iam.region := us-east-1
iam.template.file := $(pwd)/iam.cft
event.template.file := $(pwd)/event.cft

iam.export.name := $(project.name):iam:role:arn
lambda.export.name := $(project.name):lambda:function:arn

cfn.tags := Key=project,Value=$(project.name) Key=repository,Value=$(project.repo) Key=lifetime,Value=long

dist.dir := $(pwd)/.dist
dist.filename = artifact.zip
dist.file := $(dist.dir)/$(dist.filename)

build.dir := $(pwd)/.build
build.filename = main
build.file := $(build.dir)/$(build.filename)

help: ## Shows this message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Compiles the application for the Linux target (AWS Lambda)
	GOOS=linux GOARCH=amd64 go build -o $(build.file) $(project.repo)

test: ## Runs local tests
	go test $(project.repo)

require-region:
ifndef AWS_REGION
	@echo "Please specify an AWS_REGION to continue."; exit 1
endif

require-profile:
ifndef AWS_PROFILE
	@echo "Please specify an AWS_PROFILE to continue."; exit 1
endif

deploy-iam-stack: require-profile ## Deploys the IAM stack to CFN
	aws cloudformation create-stack --stack-name "$(project.name)-iam" --template-body file://$(iam.template.file) --region $(iam.region) --profile $(AWS_PROFILE) --tags $(cfn.tags) --enable-termination-protection --capabilities CAPABILITY_IAM

deploy-event-stack: require-region require-profile ## Deploys the Event stack to CFN
	aws cloudformation create-stack --stack-name $(project.name)-event --profile $(AWS_PROFILE) --template-body file://$(event.template.file) --parameters ParameterKey=IamRoleArn,ParameterValue=$(shell aws cloudformation list-exports --region $(iam.region) --query 'Exports[?Name == `$(iam.export.name)`].Value' --profile $(AWS_PROFILE) --output text) --region $(AWS_REGION) --tags $(cfn.tags) --enable-termination-protection

prepare-event-stack: require-region require-profile ## Prepares the Event stack for Go code
	aws lambda update-function-configuration --function-name $(shell aws cloudformation list-exports --region $(iam.region) --query 'Exports[?Name == `$(lambda.export.name)`].Value' --profile $(AWS_PROFILE) --output text) --region $(AWS_REGION) --handler main --runtime go1.x --profile $(AWS_PROFILE)

deploy-code: require-region require-profile ## Deploys the Go code to Lambda
	aws lambda update-function-code --function-name $(shell aws cloudformation list-exports --region $(iam.region) --query 'Exports[?Name == `$(lambda.export.name)`].Value' --profile $(AWS_PROFILE) --output text) --region $(AWS_REGION) --profile $(AWS_PROFILE) --zip-file fileb://$(dist.file)

check-iam-stack: require-profile ## Views the IAM stack status in CFN
	aws cloudformation describe-stack-events --stack-name $(project.name)-iam --region $(iam.region) --profile $(AWS_PROFILE) --query 'StackEvents[0]'

check-event-stack: require-region require-profile ## Views the Event stack status in CFN
	aws cloudformation describe-stack-events --stack-name $(project.name)-event --region $(AWS_REGION) --profile $(AWS_PROFILE) --query 'StackEvents[0]'

