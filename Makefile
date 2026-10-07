.PHONY: run test smoke vet build check

run:
	go run ./cmd/plain-default

test:
	go test ./...

smoke:
	./scripts/smoke-plain-default.sh

vet:
	go vet ./...

build:
	go build ./...

check: test vet build smoke
