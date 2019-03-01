FROM golang:1.12-alpine AS build_base

# Install tools required to build the project
RUN apk add --no-cache ca-certificates \
  curl \
  git \
  make

WORKDIR /go/src/github.com/maruware/radikocast

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

FROM build_base AS build

COPY . .

# Build the project binary
RUN go build -o /bin/radikocast ./cmd/radikocast/main.go

# This results in a single layer image
FROM alpine:latest

# Set timezone
ENV TZ "Asia/Tokyo"
# Set default output dir
VOLUME ["/output"]

RUN apk add --no-cache ca-certificates ffmpeg tzdata

COPY --from=build /bin/radikocast /bin/radikocast

ENTRYPOINT ["/bin/radikocast"]
CMD ["help"]
