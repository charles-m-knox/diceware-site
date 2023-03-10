#!/bin/sh -e

# just in case it's not in the env already
PWD="$(pwd)"

docker rm -f diceware-site || true

docker run -d \
  -p "127.0.0.1:29102:29102" \
  --restart=unless-stopped \
  --name=diceware-site \
  -v "${PWD}/cert.pem:/site/cert.pem" \
  -v "${PWD}/key.pem:/site/key.pem" \
  -w "/site" \
  -it diceware-site:latest
