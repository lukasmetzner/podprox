# Define variables
APP_NAME := podprox
CMD_DIR := ./cmd/$(APP_NAME)
MAIN_FILE := $(CMD_DIR)/main.go
BIN_DIR := ./bin
BIN_FILE := $(BIN_DIR)/$(APP_NAME)
GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')

# Define build commands
all: build

build: $(GO_FILES)
	@echo "Building $(APP_NAME)"
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_FILE) $(MAIN_FILE)
	@echo "Build complete. Binary is located at $(BIN_FILE)."

linux: $(GO_FILES)
	@echo "Building $(APP_NAME) on linux"
	@mkdir -p $(BIN_DIR)
	@GOOS=linux go build -o $(BIN_FILE) $(MAIN_FILE)
	@echo "Build complete. Binary is located at $(BIN_FILE)."

image: linux
	@echo "Building Docker Image"
	@docker build -t podprox .

k8s: k8s/podprox.yaml
	@echo "Starting k8s pod"
	@kubectl apply -f k8s/podprox.yaml

k8s-clean: k8s/podprox.yaml
	@echo "Starting k8s pod"
	@kubectl delete -f k8s/podprox.yaml

run: build
	@echo "Running $(APP_NAME)..."
	@$(BIN_FILE)

clean:
	@echo "Cleaning up..."
	@rm -rf $(BIN_DIR)
	@echo "Clean complete."

.PHONY: all build run test clean fmt vet

