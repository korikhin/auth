# syntax=docker/dockerfile:1

FROM golang:alpine3.19 AS build-stage
WORKDIR /build
COPY . .
RUN apk add --update --no-cache upx \
    && rm -rf /var/cache/apk/* \
    && CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o ./app ./cmd/auth-server \
    && upx ./app \
    && apk del upx

# Run the tests in the container
# FROM build-stage AS test-stage
# RUN go test -v ./...

FROM alpine:latest AS release-stage
WORKDIR /app
COPY --from=build-stage /build/app ./app

EXPOSE 8080
ENTRYPOINT ["./app"]
