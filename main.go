package main

import (
  "errors"
  "fmt"
  "github.com/aws/aws-sdk-go-v2/aws"
  "github.com/aws/aws-sdk-go-v2/aws/external"
  "github.com/aws/aws-sdk-go-v2/service/ec2"
  "github.com/aws/aws-lambda-go/lambda"
  "github.com/aws/aws-lambda-go/events"
  "context"
  "encoding/json"
  "strings"
)

const (
  MAX_TAG_LENGTH int = 255
)

var (
  LambdaError = errors.New("Lambda error")
  NameAlreadySetError = errors.New("Name already set")
  TagNotFoundError = errors.New("Tag not found")
  InstanceNotFoundError = errors.New("Instance not found")
  MultipleInstancesError = errors.New("Multiple instances found")
)

type AutoScalingEvent struct {
  EC2InstanceId string `json:"EC2InstanceId"`
}

func Handler(ctx context.Context, event events.CloudWatchEvent) {
  var data AutoScalingEvent
  err := json.Unmarshal(event.Detail, &data)
  if err != nil {
    panic(err.Error())
  }

  cfg, err := external.LoadDefaultAWSConfig()
  if err != nil {
    panic(err.Error())
  }

  verifyInstanceExists(cfg, data.EC2InstanceId)
  tagList := getTags(cfg, data.EC2InstanceId)

  if nameTagSet(tagList) {
    panic("Name tag already set")
  }

  projectTag, err := getTagValue(tagList, "project")
  if err != nil {
    panic(err.Error())
  }
  envTag, err := getTagValue(tagList, "environment")
  if err != nil {
    panic(err.Error())
  }

  name := buildName(projectTag, envTag, data.EC2InstanceId)
  nameInstance(cfg, data.EC2InstanceId, name)
}

func main() {
  lambda.Start(Handler)
}

func verifyInstanceExists(cfg aws.Config, instanceId string) error {
  svc := ec2.New(cfg)

  includeAll := true
  params := ec2.DescribeInstanceStatusInput{
    InstanceIds: []string{ instanceId },
    IncludeAllInstances: &includeAll,
  }

  req := svc.DescribeInstanceStatusRequest(&params)
  result, err := req.Send()
  if err != nil {
    panic(err.Error())
  }

  if len(result.InstanceStatuses) < 1 {
    fmt.Printf("Instance not found: %s\n", instanceId)
    return InstanceNotFoundError
  }

  if len(result.InstanceStatuses) > 1 {
    fmt.Printf("1+ instances found for instance %s.  Aborting to prevent damage.\n", instanceId)
    return MultipleInstancesError
  }

  return nil
}

func getTags(cfg aws.Config, instanceId string) []ec2.TagDescription {
  svc := ec2.New(cfg)

  filterName := "resource-id"
  params := ec2.DescribeTagsInput{
    Filters: []ec2.Filter {
      ec2.Filter{
        Name: &filterName,
        Values: []string{ instanceId },
      },
    },
  }

  req := svc.DescribeTagsRequest(&params)
  result, err := req.Send()
  if err != nil {
    panic(err.Error())
  }

  return result.Tags
}

func nameTagSet(tagList []ec2.TagDescription) bool {
  tagFound := false
  for _, i := range tagList {
    if *i.Key == "Name" && *i.Value != "" {
      tagFound = true
    }
  }

  return tagFound
}

func getTagValue(tagList []ec2.TagDescription, key string) (string, error) {
  var val string
  for _, i := range tagList {
    if *i.Key == key && *i.Value != "" {
      val = *i.Value
    }
  }

  if val == "" {
    return "", TagNotFoundError
  }

  return val, nil
}

func buildName(project string, env string, instanceId string) string {
  if project == "" || env == "" || instanceId == "" {
    return ""
  }
  var id string
  id_parts := strings.Split(instanceId, "-")
  if len(id_parts) == 2 {
    id = id_parts[1]
  } else {
    return ""
  }
  name := fmt.Sprintf("%s-%s-%s", project, env, id)
  if len(name) > MAX_TAG_LENGTH {
    name = name[0:MAX_TAG_LENGTH]
  }
  return name
}

func nameInstance(cfg aws.Config, instanceId string, name string) {
  svc := ec2.New(cfg)

  tagName := "Name"

  params := ec2.CreateTagsInput{
    Resources: []string{ instanceId },
    Tags: []ec2.Tag{
      ec2.Tag{
        Key: &tagName,
        Value: &name,
      },
    },
  }

  req := svc.CreateTagsRequest(&params)
  _, err := req.Send()
  if err != nil {
    // TODO do more error checking here
    panic(err.Error())
  }

  fmt.Printf("SUCCESS: Tagging %s with name %s", instanceId, name)
}

