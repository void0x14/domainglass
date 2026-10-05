# domainglass geliştirme kısayolları

BINARY := domainglass
PKG    := ./cmd/domainglass

.PHONY: all build test vet fmt clean canli kur

all: fmt vet test build

build:
	go build -o $(BINARY) $(PKG)

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

canli:
	DOMAINGLASS_CANLI_TEST=1 go test ./internal/cli/ -run Canli -v

kur: build
	install -m 0755 $(BINARY) /usr/local/bin/$(BINARY)

clean:
	go clean
	@rm -f $(BINARY)

