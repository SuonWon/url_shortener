FROM golang:1.24 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG CGO_ENABLED=0
ENV CGO_ENABLED=${CGO_ENABLED}
RUN GOOS=linux GORACH=amd64 go build -ldflags="-s -w" -o /out/app ./cmd/server

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app
COPY --from=builder /out/app /app/app

USER nonroot:nonroot
EXPOSE 8080

ENV GIN_MODE=release

ENTRYPOINT ["/app/app"]