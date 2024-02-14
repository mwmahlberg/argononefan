BINARIES = argononefan

all: $(BINARIES)

%: cmd/%/main.go fan.go temp.go
	go build ./cmd/$@

clean:
	@$(RM) $(BINARIES)

distclean: clean
	@$(RM) *.rpm