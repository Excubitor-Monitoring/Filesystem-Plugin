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
	@echo "Building frontend components"
	@make init-submodule
	@make components/Filesystems/dist
	mkdir -p frontend/
	cp components/Filesystems/dist/index.js frontend/
components/Filesystems/dist:
	@echo "Building Filesystems component"
	$(NPM) --cwd components/Filesystems/ $(NPMI)
	$(NPM) --cwd components/Filesystems/ $(NPMBUILD)
init-submodule:
	@echo "Initializing git submodule"
	git submodule update --init --recursive
build:
	make build-component
	go build -o bin/filesystem.plugin main.go
clean:
	rm -rf components/Filesystems/dist
	rm -rf bin/filesystem.plugin
	rm -rf frontend