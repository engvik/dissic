.PHONY: default
default: build

APP=dissic
VERSION=1.0.0

## build: build binaries and generate checksums
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o dist/${VERSION}/${APP}_${VERSION}_linux_amd64 cmd/dissic/main.go
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o dist/${VERSION}/${APP}_${VERSION}_darwin_amd64 cmd/dissic/main.go
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -o dist/${VERSION}/${APP}_${VERSION}_windows_amd64.exe cmd/dissic/main.go
	md5sum dist/${VERSION}/${APP}_${VERSION}_linux_amd64 > dist/${VERSION}/${APP}_${VERSION}_checksums.txt
	md5sum dist/${VERSION}/${APP}_${VERSION}_darwin_amd64 >> dist/${VERSION}/${APP}_${VERSION}_checksums.txt
	md5sum dist/${VERSION}/${APP}_${VERSION}_windows_amd64.exe >> dist/${VERSION}/${APP}_${VERSION}_checksums.txt

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
