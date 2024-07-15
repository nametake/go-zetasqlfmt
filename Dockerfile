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

FROM golang:1.22-bookworm

WORKDIR /app

COPY --from=builder /go/bin/zetasqlfmt /bin/zetasqlfmt

ENTRYPOINT ["/bin/zetasqlfmt"]
