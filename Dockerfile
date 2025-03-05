FROM golang:1-alpine AS build

WORKDIR /usr/src/selfbang

COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd

RUN go mod download

RUN go build -ldflags="-s -w" -o /usr/local/bin/selfbang cmd/selfbang/main.go

FROM golang:1-alpine

COPY --from=build /usr/local/bin/selfbang /selfbang

COPY bang.js ./
COPY public ./public
COPY index.html ./
COPY opensearch.xml ./

CMD ["/selfbang"]
