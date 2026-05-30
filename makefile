default: help

build:
	docker buildx build --platform linux/amd64 -t oidc_broker:1.0.0 --load .

test:
	go test -v ./...

clean:
	rm -rf bin/*

help:
	@echo 'Usage: make (build | test | clean)'