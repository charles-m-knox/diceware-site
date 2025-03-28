.PHONY=build

BUILDDIR=build
VER=0.2.2
FILE=diceware-site
BIN=$(BUILDDIR)/$(FILE)-v$(VER)
OUT_BIN_DIR=~/.local/bin
UNAME=$(shell go env GOOS)
ARCH=$(shell go env GOARCH)
BUILD_ENV=CGO_ENABLED=0
BUILD_FLAGS=-ldflags="-w -s -buildid= -X main.version=$(VER)" -trimpath

build-dev:
	$(BUILD_ENV) go build -v

mkbuilddir:
	mkdir -p $(BUILDDIR)

build-prod: mkbuilddir
	make build-$(UNAME)-$(ARCH)

test:
	go test -test.v -coverprofile=testcov.out ./... && \
	go tool cover -html=testcov.out

run:
	./$(BIN)

lint:
	golangci-lint run ./...

install:
	rsync -avP ./$(BIN)-$(UNAME)-$(ARCH) $(OUT_BIN_DIR)/$(FILE)
	chmod +x $(OUT_BIN_DIR)/$(FILE)

compress-prod: mkbuilddir
	rm -f $(BIN)-compressed
	upx --best -o ./$(BIN)-compressed $(BIN)

build-darwin-arm64: mkbuilddir
	$(BUILD_ENV) GOARCH=arm64 GOOS=darwin go build -v -o $(BIN)-darwin-arm64 $(BUILD_FLAGS)
	rm -f $(BIN)-darwin-arm64.xz
	xz -9 -e -T 12 -vv $(BIN)-darwin-arm64

build-darwin-amd64: mkbuilddir
	$(BUILD_ENV) GOARCH=amd64 GOOS=darwin go build -v -o $(BIN)-darwin-amd64 $(BUILD_FLAGS)
	rm -f $(BIN)-darwin-amd64.xz
	xz -9 -e -T 12 -vv $(BIN)-darwin-amd64

build-win-amd64: mkbuilddir
	$(BUILD_ENV) GOARCH=amd64 GOOS=windows go build -v -o $(BIN)-win-amd64-uncompressed $(BUILD_FLAGS)
	rm -f $(BIN)-win-amd64
	upx --best -o ./$(BIN)-win-amd64 $(BIN)-win-amd64-uncompressed

build-linux-arm64: mkbuilddir
	$(BUILD_ENV) GOARCH=arm64 GOOS=linux go build -v -o $(BIN)-linux-arm64-uncompressed $(BUILD_FLAGS)
	rm -f $(BIN)-linux-arm64
	upx --best -o ./$(BIN)-linux-arm64 $(BIN)-linux-arm64-uncompressed

build-linux-amd64: mkbuilddir
	$(BUILD_ENV) GOARCH=amd64 GOOS=linux go build -v -o $(BIN)-linux-amd64-uncompressed $(BUILD_FLAGS)
	rm -f $(BIN)-linux-amd64
	upx --best -o ./$(BIN)-linux-amd64 $(BIN)-linux-amd64-uncompressed


build-all: mkbuilddir build-linux-amd64 build-linux-arm64 build-win-amd64 build-mac-amd64 build-mac-arm64

# specific tasks

delete-builds:
	rm $(BUILDDIR)/*

gen-tls-certs:
	openssl genrsa -out key.pem 2048 && \
	openssl ecparam -genkey -name secp384r1 -out key.pem && \
	openssl req -new -x509 -sha256 -key key.pem -out cert.pem -days 3650

podman-build:
	podman build -t ghcr.io/charles-m-knox/diceware-site:latest -f containerfile .
	podman tag ghcr.io/charles-m-knox/diceware-site:latest ghcr.io/charles-m-knox/diceware-site:v$(VER)

# requires you to run 'podman login ghcr.io'
push-container-image:
	podman push ghcr.io/charles-m-knox/diceware-site:latest
	podman push ghcr.io/charles-m-knox/diceware-site:v$(VER)

podman-run:
	podman rm -f diceware-site || true
	podman run -d \
		-p "127.0.0.1:29102:29102" \
		--restart=unless-stopped \
		--name=diceware-site \
		-it diceware-site:latest
	podman logs -f diceware-site
