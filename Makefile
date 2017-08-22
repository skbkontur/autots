VERSION := $(shell git describe --always --tags --abbrev=0 | tail -c +1)
RELEASE := $(shell git describe --always --tags | awk -F- '{ if ($$2) dot="."} END { printf "%s\n",$$2}')

.PHONY: build test

default: clean test build

version:
	@echo ${VERSION}-${RELEASE}

test:
	go test

clean:
	rm -rf build

build:
	mkdir -p build/usr/bin
	go build -ldflags "-X main.version=${VERSION}-${RELEASE}" -o build/usr/bin/autots .

rpm:
	fpm -t rpm \
		-s "dir" \
		--description "Automatic timestamp injector" \
		-C ./build/ \
		--vendor "SKB Kontur" \
		--name "autots" \
		--version "${VERSION}" \
		--iteration "${RELEASE}" \
		-p build

default: build
