VERSION := $(shell grep -Eo '(\d\.\d\.\d)(-dev)?' main.go)

.PHONY: build docker release

build:
	go fmt ./...
	CGO_ENABLED=0 go build -o bin/auth .

docker:
	docker build -t moov.io/auth:$(VERSION) Dockerfile

release: docker
	CGO_ENABLED=0 go vet ./...
	CGO_ENABLED=0 go test ./...
	git tag $(VERSION)

release-push:
	git push origin $(VERSION)
