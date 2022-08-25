PHONY: all

compile:
	@echo "Compiling..."
	@go build -o ./goat ./main.go
	@echo "Compiling done."

test:
	@echo "Testing..."
	@go test -v ./... -covermode=count -coverprofile=coverage.out -coverpkg=./...
	@[ -f coverage.out ] && go tool cover -func=coverage.out -o=coverage.out
	@echo "Testing done."

format:
	@echo "Formatting..."
	@go fmt ./...
	@goimports -w ./
	@gofumpt -l -w ./
	@prettier --write "**/*.js"
	@echo "Formatting done."

clean:
	@echo "Cleaning..."
	@rm -rf ./goat
	@git clean -fdX
	@echo "Cleaning done."

all: format build test
	@echo "All done."

docs:
	@echo "Generating documentation..."
	@mkdocs build
	@echo "Generating documentation done."
