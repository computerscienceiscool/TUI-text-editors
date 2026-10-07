.PHONY: run guided-brief color-field ops-deck signal-accessible playground test smoke vet build check

run:
	go run ./cmd/plain-default

guided-brief:
	go run ./cmd/guided-brief

color-field:
	go run ./cmd/color-field

ops-deck:
	go run ./cmd/ops-deck

signal-accessible:
	go run ./cmd/signal-accessible

playground:
	go run ./cmd/playground

test:
	go test ./...

smoke:
	./scripts/smoke-plain-default.sh
	./scripts/smoke-guided-brief.sh
	./scripts/smoke-color-field.sh
	./scripts/smoke-ops-deck.sh
	./scripts/smoke-signal-accessible.sh
	./scripts/smoke-playground.sh

vet:
	go vet ./...

build:
	go build ./...

check: test vet build smoke
