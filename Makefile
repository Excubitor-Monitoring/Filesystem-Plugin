.PHONY: build-component clean build init-submodule

GO=go
GOTEST=$(GO) test
GOCOVER=$(GO) tool cover
GOMOD=$(GO) mod
GOBUILD=$(GO) build
GORUN=$(GO) run

NPM=yarn
NPMI=install
NPMBUILD=run build

build-component:
	make init-submodule
	$(NPM) --cwd components/Filesystems/ $(NPMI)
	$(NPM) --cwd components/Filesystems/ $(NPMBUILD)
	mkdir -p frontend/
	mv components/Filesystems/dist/index.js frontend/
init-submodule:
	git submodule update --init --recursive
build:
	make build-component
	go build -o bin/filesystem.plugin main.go
clean:
	rm -rf components/Filesystems/dist
	rm -rf bin/filesystem.plugin
	rm -rf frontend