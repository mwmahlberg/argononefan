BINARIES = argononefan
.PHONY: all clean distclean docker

all: $(BINARIES)

%: cmd/%/main.go fan.go temp.go
	go build ./cmd/$@

clean:
	@$(RM) $(BINARIES)

distclean: clean
	@$(RM) *.rpm

docker:
	docker build -t mwmahlberg/argononefan-rpm-builder .