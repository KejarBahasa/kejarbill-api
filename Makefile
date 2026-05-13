APP_NAME := kejarbill-api
MAIN_FILE := cmd/api/main.go

.PHONY: all build run dev test clean

all: build run

build:
	CGO_ENABLED=0 go build -mod=readonly -o ./tmp/$(APP_NAME) $(MAIN_FILE)

run:
	./tmp/$(APP_NAME)

dev:
	air

test:
	mockery --all
	go test -v ./...

clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf ./tmp/*
