package main

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"testing"
)

func TestNameTagSet(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name          string
		keyName       string
		keyVal        string
		errorExpected bool
	}{
		{"case1", "NotName", "foo", false},
		{"case2", "Name", "foo", true},
		{"case3", "Name", "", false},
	}

	for _, tc := range cases {
		tagList := []ec2.TagDescription{
			ec2.TagDescription{
				Key:   &tc.keyName,
				Value: &tc.keyVal,
			},
		}

		err := nameTagSet(tagList)
		if err != nil && !tc.errorExpected {
			t.Errorf("[%s] Expected no error", tc.Name)
		}
	}
}

func TestGetTagValue(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name          string
		keyName       string
		keyVal        string
		testKey       string
		resultVal     string
		errorExpected bool
	}{
		{"case1", "foo", "bar", "foo", "bar", false},
		{"case2", "foo", "", "foo", "", true},
		{"case3", "baz", "bar", "foo", "", true},
	}

	for _, tc := range cases {
		tagList := []ec2.TagDescription{
			ec2.TagDescription{
				Key:   &tc.keyName,
				Value: &tc.keyVal,
			},
		}

		val, err := getTagValue(tagList, tc.testKey)
		if err != nil && !tc.errorExpected {
			t.Errorf("[%s]: Expected %v but got error %v", tc.Name, tc.resultVal, err)
		}

		if val != tc.resultVal {
			t.Errorf("[%s]: Expected %v but got %v", tc.Name, tc.resultVal, val)
		}
	}
}

func TestBuildName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name       string
		project    string
		env        string
		instanceId string
		expected   string
	}{
		{"case1", "foo", "bar", "i-0123", "foo-bar-0123"},
		{"case2", "foo", "bar", "i-01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789", "foo-bar-0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456"}, // max length of tag value
		{"case3", "", "bar", "i-0123", ""},
		{"case4", "foo", "", "i-0123", ""},
		{"case5", "foo", "bar", "0123", ""},
		{"case6", "foo", "bar", "", ""},
	}

	for _, tc := range cases {
		actual := buildName(tc.project, tc.env, tc.instanceId)

		if actual != tc.expected {
			t.Errorf("[%s] Expected %v but got %v", tc.Name, tc.expected, actual)
		}
	}
}
