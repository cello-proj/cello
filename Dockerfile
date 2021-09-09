####################################################################################################
# Build binaries
####################################################################################################
FROM golang:1.16.6 as build

WORKDIR /go/src/github.com/argoproj-labs/argo-cloudops

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY internal ./internal
COPY service ./service
COPY cli ./cli
COPY Makefile ./

COPY .git ./.git

RUN make

####################################################################################################
# Run image
####################################################################################################
FROM debian:latest

RUN apt-get -yq update && \
    apt-get -yqq install ssh && \
    apt-get -yqq install openssl && \
    apt-get -yq install ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /go/src/github.com/argoproj-labs/argo-cloudops/build/service ./
COPY argo-cloudops.yaml ./

EXPOSE 8443

CMD [ "./service" ]
