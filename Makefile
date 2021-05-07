PROJECTNAME := $(shell basename "$(PWD)")

OS=linux
ARCH=amd64
OUTPUT=output

# Redirect error output to a file, so we can show it in development mode.
STDERR := /tmp/.$(PROJECTNAME)-stderr.txt

## compile: Compile the binary.
compile:
	@-touch $(STDERR)
	@-rm $(STDERR)
	@-$(MAKE) -s go-compile 2> $(STDERR)
	@cat $(STDERR) | sed -e '1s/.*/\nError:\n/'  | sed 's/make\[.*/ /' | sed "/^/s/^/     /" 1>&2

## clean: Clean build files. Runs `go clean` internally.
clean:
	@-rm -rf $(OUTPUT) 2> /dev/null
	@-$(MAKE) go-clean

go-compile: go-build

go-build: build-server build-client

build-server:
	@echo  " > Build Server binary"
	env GOOS=$(OS) GOARCH=$(ARCH) go build -o $(OUTPUT)/ezcoinrobot-server server/main.go

build-client:
	@echo  " > Build Client binary"
	env GOOS=$(OS) GOARCH=$(ARCH) go build -o $(OUTPUT)/ezcoinrobot-client client/main.go


go-clean:
	@echo "  >  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean

.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo