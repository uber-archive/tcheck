PREFIX=/usr/bin

PKG=github.com/yarpc/yab
BINNAME=$(notdir $(PKG))
CWD=$(shell pwd)

export GOPATH=$(CWD)/build
BINDIR=$(GOPATH)/bin

all: $(BINDIR)/$(BINNAME)

build:
	go build

$(GOPATH):
	mkdir -p $(GOPATH)/src
	mkdir -p $(GOPATH)/pkg
	mkdir -p $(BINDIR)

$(BINDIR)/$(BINNAME): $(GOPATH)
	git clone --depth 1 https://$(PKG) $(GOPATH)/src/$(PKG)
	cp $(GOPATH)/src/$(PKG)/man/yab.1 debian/
	cd $(GOPATH)/src/$(PKG) && glide install
	GOBIN=$(BINDIR) go install $(PKG)

install: $(BINDIR)/$(BINNAME)
	install -D $(BINDIR)/$(BINNAME) $(DESTDIR)/$(PREFIX)/$(BINNAME)

clean:
	rm -rf build $(BINNAME)

.PHONY: build install clean