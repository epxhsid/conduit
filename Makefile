.PHONY: all build clean

all: build

build:
	mkdir -p bin
	go build -o bin/conduit ./cmd

clean:
	rm -rf bin
	rmdir bin 2>/dev/null || true