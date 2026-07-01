ARG MEDIAMTX_VERSION=v1.19.2

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
ARG TARGETARCH
ARG TARGETOS

RUN apk add --no-cache git ca-certificates

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /onvif-camera ./cmd/onvif-camera/

FROM alpine:3.22 AS fetcher
ARG TARGETARCH
ARG MEDIAMTX_VERSION

RUN apk add --no-cache curl xz tar

RUN ARCH_DIR=$([ "$TARGETARCH" = "arm64" ] && echo "arm64" || echo "amd64") && \
    curl -sL "https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-${ARCH_DIR}-static.tar.xz" \
    | tar xJ -C /tmp && \
    find /tmp -name ffmpeg -type f -exec cp {} /ffmpeg \;

RUN ARCH_DIR=$([ "$TARGETARCH" = "arm64" ] && echo "arm64" || echo "amd64") && \
    curl -sL "https://github.com/bluenviron/mediamtx/releases/download/${MEDIAMTX_VERSION}/mediamtx_${MEDIAMTX_VERSION}_linux_${ARCH_DIR}.tar.gz" \
    | tar xz -C /tmp && \
    cp /tmp/mediamtx /mediamtx

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /onvif-camera /usr/local/bin/onvif-camera
COPY --from=fetcher /ffmpeg /usr/local/bin/ffmpeg
COPY --from=fetcher /mediamtx /usr/local/bin/mediamtx
EXPOSE 8080 554
ENTRYPOINT ["/usr/local/bin/onvif-camera"]
