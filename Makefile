.PHONY: run guided-brief test smoke vet build check

run:
	go run ./cmd/plain-default

guided-brief:
	go run ./cmd/guided-brief

test:
	go test ./...

smoke:
	./scripts/smoke-plain-default.sh
	./scripts/smoke-guided-brief.sh

vet:
	go vet ./...

build:
	go build ./...

check: test vet build smoke
