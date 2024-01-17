FROM docker.io/library/golang:alpine AS builder

RUN apk add upx

WORKDIR /site
COPY go.mod /site

RUN go mod download

COPY . /site

WORKDIR /site
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -v -o diceware-site-uncompressed -ldflags="-w -s -buildid=" -trimpath
RUN upx --best -o ./diceware-site diceware-site-uncompressed

FROM docker.io/library/alpine:latest
COPY --from=builder /site/diceware-site /diceware-site

CMD ["/diceware-site"]
