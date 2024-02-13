BINARIES = setfan readtemp adjustfan

all: $(BINARIES)

$(BINARIES):
	go build ./cmd/$@

install:
	chmod 755 ./deploy/install.sh
	./deploy/install.sh

uninstall:
	chmod 755 ./deploy/uninstall.sh
	./deploy/uninstall.sh

clean:
	@$(RM) $(BINARIES)