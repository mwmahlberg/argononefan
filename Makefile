BINARIES = setfan readtemp adjustfan

all: $(BINARIES)

adjustfan: cmd/adjustfan/main.go cmd/adjustfan/config.go
	go build ./cmd/$@

readtemp: cmd/readtemp/main.go
	go build ./cmd/$@

setfan: cmd/setfan/main.go
	go build ./cmd/$@

install:
	chmod 755 ./deploy/install.sh
	./deploy/install.sh

uninstall:
	chmod 755 ./deploy/uninstall.sh
	./deploy/uninstall.sh

clean:
	@$(RM) $(BINARIES)