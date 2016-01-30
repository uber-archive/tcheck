.PHONY: build install

all: build

build:
	go build

install:
	cp tcheck ${DESTDIR}/local/bin/tcheck
