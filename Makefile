.PHONY: default
default: build

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o dissic_1.0.0-beta.1_linux_amd64 cmd/dissic/main.go
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o dissic_1.0.0-beta.1_darwin_amd64 cmd/dissic/main.go
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o dissic_1.0.0-beta.1_windows_amd64.exe cmd/dissic/main.go
	md5sum dissic_1.0.0-beta.1_linux_amd64 > dissic_1.0.0-beta.1_linux_amd64.txt
	md5sum dissic_1.0.0-beta.1_darwin_amd64 > dissic_1.0.0-beta.1_darwin_amd64.txt
	md5sum dissic_1.0.0-beta.1_windows_amd64.exe > dissic_1.0.0-beta.1_windows_amd64.txt

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
