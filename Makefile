PHONY: all

build:
	@echo "Building..."
	@go build -o ./goalgotrade ./main.go
	@echo "Building done."

test:
	@echo "Testing..."
	@go test ./... -covermode=count -coverprofile=coverage.out
	@[ -f coverage.out ] && go tool cover -func=coverage.out -o=coverage.out
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
	@git clean -fdX
	@echo "Cleaning done."

all: format build test
	@echo "All done."

docs:
	@echo "Generating documentation..."
	@mkdocs build
	@echo "Generating documentation done."