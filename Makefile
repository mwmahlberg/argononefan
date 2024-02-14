BINARIES = setfan readtemp adjustfan

all: $(BINARIES)

%: cmd/%/main.go
	go build ./cmd/$@

install:
	chmod 755 ./deploy/install.sh
	./deploy/install.sh

uninstall:
	chmod 755 ./deploy/uninstall.sh
	./deploy/uninstall.sh

clean:
	@$(RM) $(BINARIES)