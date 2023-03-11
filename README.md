# diceware-site

A service that makes automatic password generation easy and secure.

## Table of contents

- [diceware-site](#diceware-site)
  - [Table of contents](#table-of-contents)
  - [Setup](#setup)
  - [Build and deploy](#build-and-deploy)
    - [Build with golang](#build-with-golang)
    - [Build with Docker](#build-with-docker)
  - [Roadmap](#roadmap)
    - [Password generation customization options](#password-generation-customization-options)
  - [Development](#development)
    - [Testing](#testing)
    - [Style](#style)
  - [Attributions](#attributions)
  - [Disclaimers](#disclaimers)

## Setup

These files come from [this repository](https://github.com/dwyl/english-words). First, extract `resources/words_alpha.tar.xz` and ensure that the file `resources/words_alpha.txt` exists. Additionally, extract `words_simple.tar.xz` and ensure that the file `resources/words_simple.txt` exists.

## Build and deploy

The diceware site is a compiled Go binary that requires a `cert.pem` and a `key.pem` in the same directory as the binary. We will provide easy steps to set this up with both Golang and Docker below.

### Build with golang

```bash
git clone https://gitea.onlinux.org/onlinux.org/diceware-site.git

cd diceware-site

go get -v

# build without cgo for compatibility on some cloud systems
# see `go tool dist list` for supported build targets
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o diceware-site
./diceware-site
```

Next, to deploy with a systemd service, the recommended default setup is as follows:

```bash
mkdir -p /opt/diceware-site
cd /opt/diceware-site
openssl genrsa -out key.pem 2048 && \
openssl ecparam -genkey -name secp384r1 -out key.pem && \
openssl req -new -x509 -sha256 -key key.pem -out cert.pem -days 3650
```

Copy `diceware-site.service` to `/etc/systemd/system/diceware-site.service` and enable it:

```bash
sudo cp diceware-site.service /etc/systemd/system/diceware-site.service
sudo systemctl daemon-reload
sudo systemctl enable --now diceware-site.service
```

Test it out:

```bash
curl -kv https://localhost:29102
```

### Build with Docker

> *Note: The Docker build process may not work yet. It is still under development.*

The Docker build process is more or less the same difficulty as the Golang build process, but does not require a systemd service.

```bash
git clone https://gitea.onlinux.org/onlinux.org/diceware-site.git

cd diceware-site

# build the image
docker build -t diceware-site:latest .
```

Now that we've built a Docker image, it doesn't matter which directory you decide to run the following steps in:

```bash
# start by generating cert.pem and key.pem
openssl genrsa -out key.pem 2048 && \
openssl ecparam -genkey -name secp384r1 -out key.pem && \
openssl req -new -x509 -sha256 -key key.pem -out cert.pem -days 3650

./run-docker.sh
```

Test it out:

```bash
curl -kv https://localhost:29102
```

## Roadmap

This application has a simple roadmap.

### Password generation customization options

Allow greater customization according to the following:

- capitalization of only first character, each word's first character, last character, or all characters
- separator for words
- easy suffix
- custom suffix

## Development

Notes that facilitate better development practices will be added here.

### Testing

Runs all tests and opens the test coverage output in the browser:

```bash
go test -test.v -coverprofile=testcov.out ./... && \
go tool cover -html=testcov.out
```

### Style

This project is crafted to operate only with the bare minimum of what is necessary to accomplish the task, within reason. For example, resources are embedded, alpinejs and a minified distribution of semantic CSS is used, only the needed routes are created, etc. While it might seem a little unconventional, the specific goal is not to just drop in a bunch of frameworks and dependencies - it's to keep everything to a minimum.

## Attributions

- Fomantic CSS
- Alpinejs
- <https://github.com/dwyl/english-words>
- <https://github.com/ulif/diceware>

## Disclaimers

This project uses the standard built-in Go cryptographic random number generation algorithms. That being said, we cannot guarantee that any password generated by this tool will keep you safe from being compromised.

Additionally, the code in this project is far from perfect. We are sharing this code with the community without any warranty. Use it at your own discretion.

This service absolutely does not store any data about its users. Frankly, we don't want your data, as it represents a liability to us.
