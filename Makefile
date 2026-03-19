BINARY     := git-tidy
BUILD_DIR  := ./bin
VERSION    ?= dev
LDFLAGS    := -ldflags "-X main.version=$(VERSION)"

.PHONY: build install clean tidy test

## build: Compile the binary into ./bin/git-tidy
build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) .
	@echo "Built $(BUILD_DIR)/$(BINARY)"

## install: Install git-tidy to /usr/local/bin so `git tidy` works
install: build
	cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed to /usr/local/bin/$(BINARY)"
	@echo "You can now run: git tidy --help"

## tidy: Run go mod tidy
tidy:
	go mod tidy

## test: Run all tests
test:
	go test ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
