package main

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func Test_NewAdmissionController(t *testing.T) {
	region := "test_region"
	bucket := "test_bucket"
	var logger *logrus.Logger
	aci, err := NewAdmissionController(region, bucket, logger)
	require.NoError(t, err)

	ac, ok := aci.(*admissionController)
	require.True(t, ok)

	require.Equal(t, ac.region, region)
	require.Equal(t, ac.bucket, bucket)
}

func Test_parseImage(t *testing.T) {
	image := "684269065708.dkr.ecr.us-east-1.amazonaws.com/stampy-webhook-admission-controller:latest"
	host, repo, tag := parseImage(image)
	require.Equal(t, host, "684269065708.dkr.ecr.us-east-1.amazonaws.com")
	require.Equal(t, repo, "stampy-webhook-admission-controller")
	require.Equal(t, tag, "latest")

	image = "684269065708.dkr.ecr.us-east-1.amazonaws.com/stampy-webhook-admission-controller/path1/path2:v1"
	host, repo, tag = parseImage(image)
	require.Equal(t, host, "684269065708.dkr.ecr.us-east-1.amazonaws.com")
	require.Equal(t, repo, "stampy-webhook-admission-controller/path1/path2")
	require.Equal(t, tag, "v1")

	image = "684269065708.dkr.ecr.us-east-1.amazonaws.com/stampy-webhook-admission-controller/path1/path2@sha256:123456"
	host, repo, tag = parseImage(image)
	require.Equal(t, host, "684269065708.dkr.ecr.us-east-1.amazonaws.com")
	require.Equal(t, repo, "stampy-webhook-admission-controller/path1/path2")
	require.Equal(t, tag, "sha256:123456")
}
