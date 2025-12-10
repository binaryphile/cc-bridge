.PHONY: build clean test

BINARY := cc-bridge
BINDIR := bin

build:
	@mkdir -p $(BINDIR)
	go build -o $(BINDIR)/$(BINARY) ./cmd/cc-bridge

clean:
	rm -rf $(BINDIR)

test:
	go test ./...
