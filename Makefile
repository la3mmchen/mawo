#!make
include build.env
export $(shell sed 's/=.*//' build.env)
GIT_COMMIT := $(shell git describe --always --long --dirty)
PROJECT_NAME := $(shell basename "$$PWD")

.DEFAULT_GOAL := default

.PHONY: default
default: fmt build run #tests

#
# *** fmt ***
#
.PHONY: fmt
fmt:
	@find . -maxdepth 1 -iname "*go" -exec go fmt {} +
	@for file in `find internal/ -iname "*go"`; do echo $$file; go fmt $$file; done;

#
# *** build steps ***
#
.PHONY: build-server
build: build-server build-cli

go-mod:
	@go mod vendor
	@go mod verify

.PHONY: build-server
build-server: go-mod
	@rm -f ${EXECUTABLE}
	@go build -o ${EXECUTABLE} -ldflags "-X main.AppVersion=${GIT_COMMIT}" .
	@echo "wrote server to ${EXECUTABLE}"

.PHONY: build-cli
build-cli: go-mod
	@rm -f ${EXECUTABLE}-cli
	@go build -o ${EXECUTABLE}-cli -ldflags "-X main.AppVersion=${GIT_COMMIT}" cli/cli.go
	@echo "wrote cli binary to ${EXECUTABLE}-cli"

.PHONY: build-image
build-image:
	docker build --build-arg app_version=$(GIT_COMMIT) -t ${PROJECT}/${PROJECT_NAME}:${GIT_COMMIT} .
	docker tag ${PROJECT}/${PROJECT_NAME}:${GIT_COMMIT} ${PROJECT}/${PROJECT_NAME}:latest

#
# *** example runs ****
#
.PHONY: run
run: run-help run-func

run-help:
	@./${EXECUTABLE} --help
	@echo "\n____________________________ \n"

run-func:
	@./${EXECUTABLE} --verbose list
	@echo "\n____________________________"

#
# *** tests ****
#
.PHONY: tests
tests: unit-tests app-tests

.PHONY: unit-tests
unit-tests:
	@go test -cover -failfast -short "."
	@echo "\n____________________________"

.PHONY: app-tests
app-tests:
	curl -Ss "localhost:8080/"
	@echo "\n____________________________"
	curl -sS "localhost:8080/health"
	@echo "\n____________________________"
	curl -sS "localhost:8080/container-resources"
	@echo "\n____________________________"
	curl -Ss "localhost:8080/container-resources?pod-label=app.kubernetes.io%2Fcomponent%3Djenkinsmaster"
	@echo "\n____________________________"
	curl -Ss "localhost:8080/container-resources?pod-label=tier%3Dcontrol-plane"
	@echo "\n____________________________"

#
# *** helm ***
#
.PHONY: deploy-dev
deploy-dev:
	set -euxo pipefail &&\
	cd mawo-helm &&\
	helm lint . &&\
	helm template . &&\
	sed -i "" "s/0.0.0/${GIT_COMMIT}/" "./Chart.yaml" &&\
	helm upgrade --namespace=${PROJECT_NAME} --create-namespace --install -f values.yaml ${PROJECT_NAME} . &&\
	kubectl --namespace=${PROJECT_NAME} rollout restart deployment ${PROJECT_NAME} &&\
	git checkout -- Chart.yaml

.PHONY: deploy-latest
deploy-latest:
	set -euxo pipefail &&\
	cd mawo-helm &&\
	helm lint . &&\
	helm template . &&\
	sed -i "" "s/0.0.0/latest/" "./Chart.yaml" &&\
	helm upgrade --namespace=${PROJECT_NAME} --create-namespace --install -f values.yaml ${PROJECT_NAME} . &&\
	kubectl --namespace=${PROJECT_NAME} rollout restart deployment ${PROJECT_NAME} &&\
	git checkout -- Chart.yaml

.PHONY: curl-on-minikube
curl-on-minikube:
	@curl -s "`minikube ip`:`kubectl -n ${PROJECT_NAME} get --no-headers svc mawo -o jsonpath='{.spec.ports[].nodePort}'`/container-resources?pod-label=tier%3Dcontrol-plane"

# minikube helper
# eval $(minikube docker-env)
# unset DOCKER_TLS_VERIFY DOCKER_HOST DOCKER_CERT_PATH"