FROM golang:1.11-stretch as builder
WORKDIR /go/src/github.com/moov-io/auth
RUN apt-get update && apt-get install make gcc g++
COPY . .
RUN make build

FROM debian:9
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=builder /go/src/github.com/moov-io/auth/bin/auth /bin/auth
EXPOSE 8080
VOLUME "/data"
ENTRYPOINT ["/bin/auth"]
