include .project/go-project.mk

GHE_HOST := git.soma.salesforce.com
SCRIPTS_PATH := ${PROJ_ROOT}/scripts

export RAPHTY_DIR=${PROJ_ROOT}
export GO111MODULE=off

.PHONY: *

.SILENT:

default: all

version:
	gofmt -r '"GIT_VERSION" -> "$(GIT_VERSION)"' version/current.template > version/current.go

all: vars gopath version build test

codecovtest: fmt vet lint
	echo "Running codecovtest"
	cd ${TEST_DIR} && go test -coverprofile=coverage.out -covermode=count ./...
	# See https://docs.codecov.io/docs/about-the-codecov-bash-uploader
	bash <(curl -s https://codecov.moe.prd-sam.prd.slb.sfdc.net/bash) -t fe599b63-00cb-45f9-9ee8-02d5d6ecd591 -f ./coverage.out -Z

build:
	echo "Running build"
	cd ${TEST_DIR} && go build -o ${PROJ_ROOT}/bin/stampy-webhook-admission-controller-aws ./