FROM golang:1.26.3-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o judo2mqtt ./cmd/judo2mqtt

FROM scratch
COPY --from=builder /build/judo2mqtt /judo2mqtt
VOLUME /config
ENTRYPOINT ["/judo2mqtt"]
