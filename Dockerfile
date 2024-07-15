FROM golang:1.22-bookworm as builder

ARG VERSION

RUN apt-get update \
  && apt-get install -y --no-install-recommends clang git \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

ENV CGO_ENABLED 1
ENV CXX clang++

WORKDIR /go-zetasqlfmt

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install ./cmd/zetasqlfmt

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    wget \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

ENV GO_VERSION 1.22.5

RUN wget https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

ENV PATH /usr/local/go/bin:$PATH

WORKDIR /app

COPY --from=builder /go/bin/zetasqlfmt /bin/zetasqlfmt

ENTRYPOINT ["/bin/zetasqlfmt"]
