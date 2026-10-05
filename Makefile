# domainglass geliştirme kısayolları

BINARY := domainglass
PKG    := ./cmd/domainglass

.PHONY: all build test vet fmt clean canli kur denetle denetle-gecmis denetle-aralik kancalar

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

denetle:
	python3 scripts/commit-denetle-test.py
	python3 scripts/commit-denetle.py --commit HEAD

denetle-gecmis:
	python3 scripts/commit-denetle.py --gecmis 20

# origin/main..HEAD aralığındaki yeni commit'leri denetler.
denetle-aralik:
	python3 scripts/commit-denetle.py --aralik origin/main..HEAD

kancalar:
	sh scripts/kur-kancalar.sh

kur: build
	install -m 0755 $(BINARY) /usr/local/bin/$(BINARY)

clean:
	go clean
	@rm -f $(BINARY)

