// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package helpers

import (
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func CompareFiles(t *testing.T, path1, path2 string) {
	t.Helper()
	path1Bs, err := ioutil.ReadFile(path1)
	require.NoError(t, err, "reading path1")

	path2Bs, err := ioutil.ReadFile(path2)
	require.NoError(t, err, "reading path2")

	require.Equal(t, string(path2Bs), string(path1Bs))
}

const BundleYAML = `---
apiVersion: imgpkg.carvel.dev/v1alpha1
kind: Bundle
metadata:
  name: my-app
authors:
- name: blah
  email: blah@blah.com
websites:
- url: blah.com
`
const ImagesYAML = `---
apiVersion: imgpkg.carvel.dev/v1alpha1
kind: ImagesLock
`
const ImageFile = "images.yml"
const BundleFile = "bundle.yml"

func ExtractDigest(t *testing.T, out string) string {
	t.Helper()
	match := regexp.MustCompile("@(sha256:[0123456789abcdef]{64})").FindStringSubmatch(out)
	require.Len(t, match, 2)
	return match[1]
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// GetDockerHubRegistry returns dockerhub registry or proxy
func GetDockerHubRegistry() string {
	dockerhubReg := "index.docker.io"
	if v, present := os.LookupEnv("DOCKERHUB_PROXY"); present {
		dockerhubReg = v
	}
	return dockerhubReg
}

// CompleteImageRef returns image reference
func CompleteImageRef(ref string) string {
	return GetDockerHubRegistry() + "/" + ref
}

// TestExpectedRegistry tests an expected
// value of DOCKERHUB_PROXY
func TestExpectedRegistry(t *testing.T) {
	origProxyVal := ""

	v, isSet := os.LookupEnv("DOCKERHUB_PROXY")
	if isSet {
		origProxyVal = v
	}

	defer os.Setenv("DOCKERHUB_PROXY", origProxyVal)

	os.Unsetenv("DOCKERHUB_PROXY")
	assert.Equal(t, "index.docker.io", GetDockerHubRegistry())

	os.Setenv("DOCKERHUB_PROXY", "my-dockerhub-proxy.tld/dockerhub-proxy")
	assert.Equal(t, "my-dockerhub-proxy.tld/dockerhub-proxy", GetDockerHubRegistry())
	os.Unsetenv("DOCKERHUB_PROXY")
}

// TestExpectedImgRef tests an expected
// value of image reference
func TestExpectedImgRef(t *testing.T) {
	origProxyVal := ""

	v, isSet := os.LookupEnv("DOCKERHUB_PROXY")
	if isSet {
		origProxyVal = v
	}

	defer os.Setenv("DOCKERHUB_PROXY", origProxyVal)

	os.Unsetenv("DOCKERHUB_PROXY")
	assert.Equal(t,
		"index.docker.io/library/hello-world@sha256:ebf526c198a14fa138634b9746c50ec38077ec9b3986227e79eb837d26f59dc6",
		CompleteImageRef("library/hello-world@sha256:ebf526c198a14fa138634b9746c50ec38077ec9b3986227e79eb837d26f59dc6"))

	os.Setenv("DOCKERHUB_PROXY", "my-dockerhub-proxy.tld/dockerhub-proxy")
	assert.Equal(t,
		"my-dockerhub-proxy.tld/dockerhub-proxy/library/hello-world@sha256:ebf526c198a14fa138634b9746c50ec38077ec9b3986227e79eb837d26f59dc6",
		CompleteImageRef("library/hello-world@sha256:ebf526c198a14fa138634b9746c50ec38077ec9b3986227e79eb837d26f59dc6"))
	os.Unsetenv("DOCKERHUB_PROXY")
}
