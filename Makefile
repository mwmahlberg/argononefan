BINARIES = argononefan
CMDSOURCES = $(wildcard cmd/argononefan/*.go)
SOURCES = $(CMDSOURCES) fan.go temperature.go
GOLDFLAGS := ${GOLDFLAGS} -X main.version=$(shell git describe --tags --no-always --dirty)
.PHONY: all clean distclean docker

all: $(BINARIES)

argononefan: $(SOURCES)
	go build -ldflags "${GOLDFLAGS}" ./cmd/$@

clean:
	@$(RM) $(BINARIES)

distclean: clean
	@$(RM) *.rpm

docker:
	docker build -t mwmahlberg/argononefan-rpm-builder .