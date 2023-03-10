FROM golang:alpine AS builder

WORKDIR /site
COPY go.mod /site

RUN go mod download

COPY . /site

WORKDIR /site
RUN go build -v

FROM alpine:latest
COPY --from=builder /site/diceware-site /diceware-site

CMD ["/diceware-site"]
