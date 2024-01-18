.PHONY=build

BUILDDIR=build
VER=0.1.2
BIN=$(BUILDDIR)/diceware-site-v$(VER)

build-dev:
	CGO_ENABLED=0 go build -v

mkbuilddir:
	mkdir -p $(BUILDDIR)

build-prod: mkbuilddir
	CGO_ENABLED=0 go build -v -o $(BIN) -ldflags="-w -s -buildid=" -trimpath

test:
	go test -test.v -coverprofile=testcov.out ./... && \
	go tool cover -html=testcov.out

run:
	./$(BIN)

lint:
	golangci-lint run ./...

compress-prod: mkbuilddir
	rm -f $(BIN)-compressed
	upx --best -o ./$(BIN)-compressed $(BIN)

build-mac-arm64: mkbuilddir
	CGO_ENABLED=0 GOARCH=arm64 GOOS=darwin go build -v -o $(BIN)-darwin-arm64 -ldflags="-w -s -buildid=" -trimpath
	rm -f $(BIN)-darwin-arm64.xz
	xz -9 -e -T 12 -vv $(BIN)-darwin-arm64

build-mac-amd64: mkbuilddir
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -v -o $(BIN)-darwin-amd64 -ldflags="-w -s -buildid=" -trimpath
	rm -f $(BIN)-darwin-amd64.xz
	xz -9 -e -T 12 -vv $(BIN)-darwin-amd64

build-win-amd64: mkbuilddir
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -v -o $(BIN)-win-amd64-uncompressed -ldflags="-w -s -buildid=" -trimpath
	rm -f $(BIN)-win-amd64
	upx --best -o ./$(BIN)-win-amd64 $(BIN)-win-amd64-uncompressed

build-linux-arm64: mkbuilddir
	CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -v -o $(BIN)-linux-arm64-uncompressed -ldflags="-w -s -buildid=" -trimpath
	rm -f $(BIN)-linux-arm64
	upx --best -o ./$(BIN)-linux-arm64 $(BIN)-linux-arm64-uncompressed

build-linux-amd64: mkbuilddir
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -v -o $(BIN)-linux-amd64-uncompressed -ldflags="-w -s -buildid=" -trimpath
	rm -f $(BIN)-linux-amd64
	upx --best -o ./$(BIN)-linux-amd64 $(BIN)-linux-amd64-uncompressed

build-all: mkbuilddir build-linux-amd64 build-linux-arm64 build-win-amd64 build-mac-amd64 build-mac-arm64

delete-builds:
	rm $(BUILDDIR)/*

gen-tls-certs:
	openssl genrsa -out key.pem 2048 && \
	openssl ecparam -genkey -name secp384r1 -out key.pem && \
	openssl req -new -x509 -sha256 -key key.pem -out cert.pem -days 3650

podman-build:
	podman build -t gitea.cmcode.dev/cmcode/diceware-site:latest -f containerfile .
	podman tag gitea.cmcode.dev/cmcode/diceware-site:latest gitea.cmcode.dev/cmcode/diceware-site:v$(VER)

# requires you to run 'podman login gitea.cmcode.dev'
push-gitea-container-image:
	podman push gitea.cmcode.dev/cmcode/diceware-site:latest
	podman push gitea.cmcode.dev/cmcode/diceware-site:v$(VER)

podman-run:
	podman rm -f diceware-site || true
	podman run -d \
		-p "127.0.0.1:29102:29102" \
		--restart=unless-stopped \
		--name=diceware-site \
		-it diceware-site:latest
	podman logs -f diceware-site
