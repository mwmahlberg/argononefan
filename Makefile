BINARIES = argononefan
.PHONY: all clean distclean docker

all: $(BINARIES)

%: cmd/%/main.go cmd/%/help.go cmd/%/thresholds.go fan.go temperature.go
	go build ${GOLDFLAGS} ./cmd/$@

clean:
	@$(RM) $(BINARIES)

distclean: clean
	@$(RM) *.rpm

docker:
	docker build -t mwmahlberg/argononefan-rpm-builder .