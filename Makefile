PHONY: all

compile:
	@echo "Compiling..."
	@GOARCH=amd64 go build -v -o ./goat-amd64 ./main.go
	@GOARCH=arm64 go build -v -o ./goat-arm64 ./main.go
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

all: format compile test
	@echo "All done."

docs:
	@echo "Generating documentation..."
	@mkdocs build
	@echo "Generating documentation done."
