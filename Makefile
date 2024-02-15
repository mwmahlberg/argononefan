BINARIES = argononefan
.PHONY: all clean distclean docker

all: $(BINARIES)

%: cmd/%/main.go fan.go temperature.go
	go build ./cmd/$@

clean:
	@$(RM) $(BINARIES)

distclean: clean
	@$(RM) *.rpm

docker:
	docker build -t mwmahlberg/argononefan-rpm-builder .