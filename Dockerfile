ARG GO_VERSION
ARG GOARCH
ARG BASE_IMAGE

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build -o ipruler cmd/main.go

FROM ${BASE_IMAGE}

WORKDIR /app

COPY --from=builder /app/ipruler /app/ipruler

ENTRYPOINT ["/app/ipruler"]