VERSION := $(shell grep -Eo '(v[0-9]+[\.][0-9]+[\.][0-9]+(-dev)?)' main.go)

.PHONY: build docker release

build:
	go fmt ./...
	CGO_ENABLED=1 go build -o bin/auth .

docker:
	docker build -t moov.io/auth:$(VERSION) -f Dockerfile .

release: docker
	CGO_ENABLED=0 go vet ./...
	CGO_ENABLED=0 go test ./...
	git tag $(VERSION)

release-push:
	git push origin $(VERSION)
