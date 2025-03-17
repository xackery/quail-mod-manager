NAME ?= quail-mod-manager
VERSION ?= 0.1
EQ_PATH := ../eq/rof2

run: build-windows
	@echo "run: running"
	mkdir -p bin
	cp ${EQ_PATH}/dodequip.eqg bin/
	cd bin && wine ${NAME}.exe

run-%: build-windows
	@echo "run: running"
	mkdir -p bin
	@cp ${EQ_PATH}/$* bin/

	cd bin && wine64 ${NAME}.exe $*

view-quail-%: build-windows
	cd bin && wine64 ${NAME}.exe


# CICD triggers this
.PHONY: set-variable
set-version:
	@go mod edit -dropreplace github.com/xackery/wlk
	@go mod edit -dropreplace github.com/xackery/quail
	@echo "VERSION=${VERSION}" >> $$GITHUB_ENV


#go install golang.org/x/tools/cmd/goimports@latest
#go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
#go install golang.org/x/lint/golint@latest
#go install honnef.co/go/tools/cmd/staticcheck@v0.2.2

sanitize:
	@echo "sanitize: checking for errors"
	rm -rf vendor/
	go vet -tags ci ./...
	test -z $(goimports -e -d . | tee /dev/stderr)
	gocyclo -over 30 .
	golint -set_exit_status $(go list -tags ci ./...)
	staticcheck -go 1.14 ./...
	go test -tags ci -covermode=atomic -coverprofile=coverage.out ./...
    coverage=`go tool cover -func coverage.out | grep total | tr -s '\t' | cut -f 3 | grep -o '[^%]*'`

.PHONY: build-all
build-all: sanitize build-prepare build-linux build-darwin build-windows
.PHONY: build-prepare
build-prepare:
	@echo "Preparing talkeq ${VERSION}"
	@rm -rf bin/*
	@-mkdir -p bin/
.PHONY: build-darwin
build-darwin:
	@echo "Building darwin ${VERSION}"
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -s -w" -o bin/${NAME}-darwin-x64 main.go
.PHONY: build-linux
build-linux:
	@echo "Building Linux ${VERSION}"
	@GOOS=linux GOARCH=amd64 go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -w" -o bin/${NAME}-linux-x64 main.go
.PHONY: build-windows
build-windows:
	@echo "Building Windows ${VERSION}"
	mkdir -p bin
	go install github.com/akavel/rsrc@latest
	rsrc -ico quail-mod-manager.ico -manifest quail-mod-manager.exe.manifest
	cp quail-mod-manager.exe.manifest bin/
	GOOS=windows GOARCH=amd64 go build -buildmode=pie -ldflags="-X main.Version=${VERSION} -s -w -H=windowsgui" -o bin/${NAME}.exe