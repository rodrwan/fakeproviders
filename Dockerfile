FROM alpine

RUN apk add --update ca-certificates

COPY bin/fakeprovider /usr/bin/fakeprovider

EXPOSE 9005

ENTRYPOINT [ "fakeprovider" ]
