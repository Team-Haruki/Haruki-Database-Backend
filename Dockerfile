FROM golang:1.25-alpine AS builder
ARG VERSION=dev
ARG GIT_SHA=unknown
ARG BUILD_DATE=unknown
WORKDIR /app
COPY . .
RUN apk add --no-cache build-base
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build \
    -trimpath \
    -ldflags="-s -w \
      -X 'main.Version=${VERSION}' \
      -X 'main.Commit=${GIT_SHA}' \
      -X 'main.BuildDate=${BUILD_DATE}'" \
    -o haruki-database-backend ./main.go

FROM alpine:3.20

ARG VERSION=dev
ARG GIT_SHA=unknown
ARG BUILD_DATE=unknown
LABEL org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.revision=$GIT_SHA \
      org.opencontainers.image.created=$BUILD_DATE

WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/haruki-database-backend .

EXPOSE 6666
ENTRYPOINT ["./haruki-database-backend"]