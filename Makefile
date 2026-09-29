VERSION ?= 0.1.0
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test vet demo install uninstall clean

build:
	GOTOOLCHAIN=auto go build -ldflags "$(LDFLAGS)" -o hazyvpn .

test:
	GOTOOLCHAIN=auto go test ./...

vet:
	GOTOOLCHAIN=auto go vet ./...

demo: build
	./hazyvpn --demo

install:
	HAZYVPN_VERSION=$(VERSION) ./install.sh

uninstall:
	./uninstall.sh

clean:
	rm -f hazyvpn
