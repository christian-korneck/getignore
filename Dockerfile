FROM golang:1.16-alpine AS builder
WORKDIR /go/src/github.com/christian-korneck/getignore
COPY . .
RUN apk add --no-cache ca-certificates && update-ca-certificates
ENV CGO_ENABLED 0
RUN go build -v -a -tags netgo -ldflags='-s -w -extldflags "-static"' .

FROM scratch
COPY --from=builder /go/src/github.com/christian-korneck/getignore/getignore /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

#nobody:nogroup
USER 65534:65534
ENTRYPOINT ["/getignore"]