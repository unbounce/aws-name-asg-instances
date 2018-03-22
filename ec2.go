package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"strings"
)

const (
	MAX_TAG_LENGTH int = 255
)

var (
	NameAlreadySetError    = errors.New("Name already set")
	TagNotFoundError       = errors.New("Tag not found")
	InstanceNotFoundError  = errors.New("Instance not found")
	MultipleInstancesError = errors.New("Multiple instances found")
)

func verifyInstanceExists(cfg aws.Config, instanceId string) error {
	svc := ec2.New(cfg)

	includeAll := true
	params := ec2.DescribeInstanceStatusInput{
		InstanceIds:         []string{instanceId},
		IncludeAllInstances: &includeAll,
	}

	req := svc.DescribeInstanceStatusRequest(&params)
	result, err := req.Send()
	if err != nil {
		panic(err.Error())
	}

	if len(result.InstanceStatuses) < 1 {
		return InstanceNotFoundError
	}

	if len(result.InstanceStatuses) > 1 {
		return MultipleInstancesError
	}

	return nil
}

func getTags(cfg aws.Config, instanceId string) []ec2.TagDescription {
	svc := ec2.New(cfg)

	filterName := "resource-id"
	params := ec2.DescribeTagsInput{
		Filters: []ec2.Filter{
			ec2.Filter{
				Name:   &filterName,
				Values: []string{instanceId},
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

func nameTagSet(tagList []ec2.TagDescription) error {
	for _, i := range tagList {
		if *i.Key == "Name" && *i.Value != "" {
			return NameAlreadySetError
		}
	}

	return nil
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
		Resources: []string{instanceId},
		Tags: []ec2.Tag{
			ec2.Tag{
				Key:   &tagName,
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
}
