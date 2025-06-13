.PHONY: start

#start:
#	go run cmd/main.go
PROJECT?=github.com/ivanbulyk/vortexq
APP?=vortexq
PORT?=8085

RELEASE?=0.0.1
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
CONTAINER_IMAGE?=stackyogi/${APP}

HOST?=0.0.0.0



clean:
	rm -f ${APP}

build: clean
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
             		-ldflags "-s -w -X ${PROJECT}/internal/version.Release=${RELEASE} \
             		-X ${PROJECT}/internal/version.Commit=${COMMIT} -X ${PROJECT}/internal/version.BuildTime=${BUILD_TIME}" \
             		-o ./bin/${APP} cmd/main.go

container: build
	docker build --build-arg COMMIT=$(COMMIT) --build-arg BUILD_TIME=$(BUILD_TIME) --build-arg RELEASE=$(RELEASE) \
     		--build-arg PROJECT=$(PROJECT) --build-arg SERVER_SERVICE_HOST=$(HOST) -t $(CONTAINER_IMAGE):$(RELEASE) .


run: container
	docker stop $(APP):$(RELEASE) || true && docker rm $(APP):$(RELEASE) || true
	docker run --name ${APP} -p ${PORT}:${PORT} --rm \
		-e "PORT=${PORT}" \
		stackyogi/$(APP):$(RELEASE)

start: build
	PORT=${PORT} ./bin/${APP}

test:
	go test -v -race ./...

push: container
	docker push $(CONTAINER_IMAGE):$(RELEASE)

.DEFAULT_GOAL := start