FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build_base

# Install tools required to build the project
RUN apk add --no-cache ca-certificates \
  curl \
  git \
  make

WORKDIR /go/src/github.com/maruware/radikocast

ENV GO111MODULE=on

ARG TARGETOS
ARG TARGETARCH

COPY go.mod .
COPY go.sum .

RUN go mod download

FROM build_base AS build

COPY . .

# Build the project binary with explicit target platform to avoid qemu compile instability.
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /bin/radikocast ./cmd/radikocast/main.go

# This results in a single layer image
FROM alpine:latest

# Set timezone
ENV TZ="Asia/Tokyo"
# Set default output dir
VOLUME ["/output"]

RUN apk add --no-cache ca-certificates ffmpeg tzdata

COPY --from=build /bin/radikocast /bin/radikocast

ENTRYPOINT ["/bin/radikocast"]
CMD ["help"]
