package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
)

const (
	TAG_PROJECT    string = "project"
	TAG_ENV        string = "environment"
)

type AutoScalingEvent struct {
	EC2InstanceId string `json:"EC2InstanceId"`
}

func processPayload(event events.CloudWatchEvent) AutoScalingEvent {
	var data AutoScalingEvent
	err := json.Unmarshal(event.Detail, &data)
	if err != nil {
		panic(err.Error()) // AWS changed their payload.  Fail fast.
	}
  return data
}

func Handler(ctx context.Context, event events.CloudWatchEvent) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic(err.Error()) // Lambda or IAM is failing.  Fail fast.
	}

  data := processPayload(event)

  if err := verifyInstanceExists(cfg, data.EC2InstanceId); err != nil {
    switch err {
      case InstanceNotFoundError:
        fmt.Printf("Instance not found: %s\n", data.EC2InstanceId)
        return
      case MultipleInstancesError:
        fmt.Printf("1+ instances found for instance %s.  Aborting to prevent damage.\n", data.EC2InstanceId)
        return
      default:
        panic(err.Error())
    }
  }

	tagList := getTags(cfg, data.EC2InstanceId)

	if err := nameTagSet(tagList); err != nil {
		fmt.Printf("Name tag already set on instance: %s\n", data.EC2InstanceId)
		return
	}

	projectTag, err := getTagValue(tagList, TAG_PROJECT)
	if err != nil {
		fmt.Printf("Missing required tag: %s\n", TAG_PROJECT)
		return
	}
	envTag, err := getTagValue(tagList, TAG_ENV)
	if err != nil {
		fmt.Printf("Missing required tag: %s\n", TAG_ENV)
		return
	}

	name := buildName(projectTag, envTag, data.EC2InstanceId)
	nameInstance(cfg, data.EC2InstanceId, name)
	fmt.Printf("SUCCESS: Tagging %s with name %s", data.EC2InstanceId, name)
}

func main() {
	lambda.Start(Handler)
}

