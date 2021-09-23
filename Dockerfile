FROM debian:latest

ARG BINARY

RUN apt-get -yq update && \
    apt-get -yq install ssh openssl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY $BINARY argo-cloudops.yaml ./

EXPOSE 8443

CMD [ "./service" ]
