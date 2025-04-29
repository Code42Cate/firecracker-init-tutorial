.PHONY: build setup run

build:
	@echo "Building Go binary and setting up filesystem..."
	./fs.sh

setup:
	@echo "Running setup script..."
	./setup.sh

run:
	@echo "Running Go program..."
	go run main.go 