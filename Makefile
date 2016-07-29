PREFIX=/usr/bin

PKG=github.com/uber/tcheck
BINNAME=$(notdir $(PKG))
CWD=$(shell pwd)

export GOPATH=$(CWD)/build
BINDIR=$(GOPATH)/bin

.PHONY: all
all: $(BINDIR)/$(BINNAME)

.PHONY: build
build:
	go build

$(GOPATH):
	mkdir -p $(GOPATH)/src
	mkdir -p $(GOPATH)/pkg
	mkdir -p $(BINDIR)

$(BINDIR)/$(BINNAME): $(GOPATH)
	git clone --depth 1 https://$(PKG) $(GOPATH)/src/$(PKG)
	cp $(GOPATH)/src/$(PKG)/man/tcheck.1 debian/
	cd $(GOPATH)/src/$(PKG) && glide install
	GOBIN=$(BINDIR) go install $(PKG)

.PHONY: install
install: $(BINDIR)/$(BINNAME)
	install -D $(BINDIR)/$(BINNAME) $(DESTDIR)/$(PREFIX)/$(BINNAME)

.PHONY: clean
clean:
	rm -rf build $(BINNAME)
