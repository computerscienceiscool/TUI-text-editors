.PHONY: run guided-brief color-field test smoke vet build check

run:
	go run ./cmd/plain-default

guided-brief:
	go run ./cmd/guided-brief

color-field:
	go run ./cmd/color-field

test:
	go test ./...

smoke:
	./scripts/smoke-plain-default.sh
	./scripts/smoke-guided-brief.sh
	./scripts/smoke-color-field.sh

vet:
	go vet ./...

build:
	go build ./...

check: test vet build smoke
