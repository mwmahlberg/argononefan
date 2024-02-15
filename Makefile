BINARIES = argononefan
CMDSOURCES = $(wildcard cmd/argononefan/*.go)
SOURCES = $(CMDSOURCES) fan.go temperature.go
.PHONY: all clean distclean docker

all: $(BINARIES)

argononefan: $(SOURCES)
	go build ${GOLDFLAGS} ./cmd/$@

clean:
	@$(RM) $(BINARIES)

distclean: clean
	@$(RM) *.rpm

docker:
	docker build -t mwmahlberg/argononefan-rpm-builder .