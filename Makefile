include .project/go-project.mk

GHE_HOST := git.soma.salesforce.com
SCRIPTS_PATH := ${PROJ_ROOT}/scripts

export RAPHTY_DIR=${PROJ_ROOT}
export GO111MODULE=off

.PHONY: *

.SILENT:

default: all

all: vars gopath vendor build

build:
	echo "Running build"
	cd ${TEST_DIR} && go build -o ${PROJ_ROOT}/bin/stampy-admission-webhook ./

vendor: get

get:
	echo "*** GOPATH=${GOPATH}"
	$(call gitclone,${GHE_HOST},go-mirrors/glog,                                    ${GOPATH}/src/github.com/golang/glog,                              23def4e6c14b4da8ac2ed8007337bc5eb5007998)
	$(call gitclone,${GHE_HOST},go-mirrors/ghodss-yaml,                             ${GOPATH}/src/github.com/ghodss/yaml,                              25d852aebe32c875e9c044af3eef9c7dc6bc777f)
	$(call gitclone,${GHE_HOST},go-mirrors/gogo-protobuf,                           ${GOPATH}/src/github.com/gogo/protobuf,                            88dda4156dab6c722c91df813dd73f94253b09b6)
	$(call gitclone,${GHE_HOST},go-mirrors/spf13-pflag,                             ${GOPATH}/src/github.com/spf13/pflag,                              24fa6976df40757dce6aea913e7b81ade90530e1)
	$(call gitclone,${GHE_HOST},go-mirrors/golang-net,                              ${GOPATH}/src/golang.org/x/net,                                    66aacef3dd8a676686c7ae3716979581e8b03c47)
	$(call gitclone,${GHE_HOST},go-mirrors/golang-text,                             ${GOPATH}/src/golang.org/x/text,                                   b19bf474d317b857955b12035d2c5acb57ce8b01)
	$(call httpsclone,${GITHUB_HOST},go-yaml/yaml,                                  ${GOPATH}/src/gopkg.in/yaml.v2,                                    7b8349ac747c6a24702b762d2c4fd9266cf4f1d6)
	$(call httpsclone,${GITHUB_HOST},docker/distribution,                           ${GOPATH}/src/docker/distribution,                                 6d62eb1d4a3515399431b713fde3ce5a9b40e8d5)
	$(call httpsclone,${GITHUB_HOST},json-iterator/go,                              ${GOPATH}/src/github.com/json-iterator/go,                         f2b4162afba35581b6d4a50d3b8f34e33c144682)
	$(call httpsclone,${GITHUB_HOST},kubernetes/kubernetes,                         ${GOPATH}/src/k8s.io/kubernetes,                                   release-1.10)
	$(call httpsclone,${GITHUB_HOST},kubernetes/api,                                ${GOPATH}/src/k8s.io/api,                                          release-1.10)
	$(call httpsclone,${GITHUB_HOST},kubernetes/apimachinery,                       ${GOPATH}/src/k8s.io/apimachinery,                                 release-1.10)
	$(call httpsclone,${GITHUB_HOST},kubernetes/apiextensions-apiserver,            ${GOPATH}/src/k8s.io/apiextensions-apiserver,                      release-1.10)
	$(call httpsclone,${GITHUB_HOST},kubernetes/apiserver,                          ${GOPATH}/src/k8s.io/apiserver,                                    release-1.10)
	$(call httpsclone,${GITHUB_HOST},kubernetes/client-go,                          ${GOPATH}/src/k8s.io/client-go,                                    release-11.0)
	$(call httpsclone,${GITHUB_HOST},google/gofuzz,                                 ${GOPATH}/src/github.com/google/gofuzz,                            24818f796faf91cd76ec7bddd72458fbced7a6c1)
	$(call httpsclone,${GITHUB_HOST},modern-go/concurrent,                          ${GOPATH}/src/github.com/modern-go/concurrent,                     bacd9c7ef1dd9b15be4a9909b8ac7a4e313eec94)
	$(call httpsclone,${GITHUB_HOST},modern-go/reflect2,                            ${GOPATH}/src/github.com/modern-go/reflect2,                       94122c33edd36123c84d5368cfb2b69df93a0ec8)
	$(call httpsclone,${GITHUB_HOST},go-inf/inf,                                    ${GOPATH}/src/gopkg.in/inf.v0,                                     8237a9a5367b2a82f922b38d4b3676293e031763)

docker: build
	CGO_ENABLED=0 GOOS=linux docker build --no-cache -t ${DOCKER_USER}/stampy-admission-webhook:v1 .
	docker push ${DOCKER_USER}/stampy-admission-webhook:v1