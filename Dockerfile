FROM golang:1.11-alpine as builder
WORKDIR /go/src/github.com/moov-io/auth
RUN apk -U add make gcc
COPY . .
RUN make build

FROM scratch
COPY --from=builder /go/src/github.com/moov-io/auth/bin/auth /bin/auth
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
EXPOSE 8080
ENTRYPOINT ["/bin/auth"]
VOLUME "/data"
