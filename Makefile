BINARIES = setfan argononefan

all: $(BINARIES)

%: cmd/%/main.go
	go build ./cmd/$@

clean:
	@$(RM) $(BINARIES)