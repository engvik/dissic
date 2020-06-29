.PHONY: default
default: help

.PHONY: run
## run: Run dissic (set CONFIG=path/to/config.yaml)
run:
	go run -race cmd/dissic/main.go --config=$(CONFIG)

.PHONY: test
## test: Run the tests
test:
	go test -race -v ./...

.PHONY: help
## help: Print this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
