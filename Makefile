PHONY: all

build:
	@echo "Building..."
	@go build -o ./goalgotrade ./main.go
	@echo "Building done."

test:
	@echo "Testing..."
	@go test ./...
	@echo "Testing done."

all: build test
	@echo "All done."
