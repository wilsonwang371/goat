PHONY: all

build:
	@echo "Building..."
	@go build -o ./goalgotrade ./main.go
	@echo "Building done."

test:
	@echo "Testing..."
	@go test ./...
	@echo "Testing done."

format:
	@echo "Formatting..."
	@go fmt ./...
	@goimports -w ./
	@gofumpt -l -w ./
	@echo "Formatting done."

clean:
	@echo "Cleaning..."
	@rm -rf ./goalgotrade
	@echo "Cleaning done."

all: format build test
	@echo "All done."
