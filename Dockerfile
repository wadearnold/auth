# build stage
FROM golang:alpine AS build-env
ADD . /src
RUN cd /src && go build -o auth

# final stage
FROM alpine
WORKDIR /moov
COPY --from=build-env /src/auth /moov/
ENTRYPOINT ./auth
EXPOSE 8080